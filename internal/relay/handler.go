package relay

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/pkg/protocol"
)

type AgentIdentity struct {
	AgentID        string
	OrganizationID string
	MaxTunnels     int
	MaxConnections int
	BandwidthBytes int64
}

type AgentAuthenticator interface {
	Authenticate(context.Context, string) (AgentIdentity, error)
}

type RelayAffinity interface {
	Claim(context.Context, string, string, time.Duration) (bool, error)
	Release(context.Context, string, string) error
}

type ManagedTunnelResolver interface {
	Resolve(context.Context, string) (ManagedTunnelPolicy, error)
}

type ManagedTunnelPasswordVerifier interface {
	VerifyPassword(context.Context, string, string) (bool, error)
}

type ManagedTunnelPolicy struct {
	OrganizationID    string
	PublicHostname    string
	PasswordHash      string
	PasswordProtected bool
	Status            string
}

type connectionState struct {
	negotiated    bool
	authenticated bool
	identity      AgentIdentity
}

type Handler struct {
	authenticator  AgentAuthenticator
	sessions       *engine.SessionRegistry
	router         *engine.RequestRouter
	tcp            *TCPManager
	udp            *UDPManager
	maxSessions    int
	maxTunnels     int
	maxBandwidth   int64
	heartbeat      time.Duration
	readTimeout    time.Duration
	maxFrameBytes  int64
	drainTimeout   time.Duration
	logger         *slog.Logger
	metrics        *Metrics
	bandwidth      *engine.BandwidthLimiter
	usage          engine.UsageRecorder
	allowedOrigins []string
	publicDomain   string
	affinity       RelayAffinity
	managedTunnels ManagedTunnelResolver
	relayID        string
	affinityTTL    time.Duration
	mu             sync.Mutex
	connections    int
	orgLimits      map[string]int
	orgConnections map[string]int
	writeMu        sync.Mutex
}

type HandlerOptions struct {
	MaxConnections int
	MaxTunnels     int
	MaxBandwidth   int64
	Heartbeat      time.Duration
	ReadTimeout    time.Duration
	MaxFrameBytes  int64
	DrainTimeout   time.Duration
	Logger         *slog.Logger
	Metrics        *Metrics
	UsageRecorder  engine.UsageRecorder
	AllowedOrigins string
	PublicDomain   string
	Affinity       RelayAffinity
	RelayID        string
	AffinityTTL    time.Duration
	ManagedTunnels ManagedTunnelResolver
}

func NewHandler(authenticator AgentAuthenticator, sessions *engine.SessionRegistry, router *engine.RequestRouter, tcp *TCPManager, udp *UDPManager, maxSessions int) (*Handler, error) {
	return NewHandlerWithOptions(authenticator, sessions, router, tcp, udp, HandlerOptions{MaxConnections: maxSessions, MaxTunnels: maxSessions, Heartbeat: 20 * time.Second, ReadTimeout: 90 * time.Second, MaxFrameBytes: 16 << 20, PublicDomain: "tunnel.outpipe.dev"})
}

func NewHandlerWithOptions(authenticator AgentAuthenticator, sessions *engine.SessionRegistry, router *engine.RequestRouter, tcp *TCPManager, udp *UDPManager, options HandlerOptions) (*Handler, error) {

	if authenticator == nil || sessions == nil || router == nil || options.MaxConnections < 1 || options.MaxTunnels < 1 || options.Heartbeat <= 0 || options.ReadTimeout <= options.Heartbeat || options.MaxFrameBytes < 1 {
		return nil, fmt.Errorf("authenticator, session registry, router, and positive session limit are required")
	}

	if tcp == nil || udp == nil {
		return nil, fmt.Errorf("tcp and udp managers are required")
	}

	if options.Logger == nil {
		options.Logger = slog.Default()
	}

	if options.Metrics == nil {
		options.Metrics = NewMetrics()
	}

	if options.DrainTimeout < 0 {
		return nil, fmt.Errorf("drain timeout cannot be negative")
	}

	if options.Affinity != nil && (options.RelayID == "" || options.AffinityTTL <= 0) {
		return nil, fmt.Errorf("relay affinity requires relay id and positive ttl")
	}

	if strings.TrimSpace(options.PublicDomain) == "" {
		return nil, fmt.Errorf("public tunnel domain is required")
	}

	tcp.SetMaxConnections(options.MaxConnections)
	udp.SetMaxPackets(options.MaxConnections)
	handler := &Handler{authenticator: authenticator, sessions: sessions, router: router, tcp: tcp, udp: udp, maxSessions: options.MaxConnections, maxTunnels: options.MaxTunnels, maxBandwidth: options.MaxBandwidth, heartbeat: options.Heartbeat, readTimeout: options.ReadTimeout, maxFrameBytes: options.MaxFrameBytes, drainTimeout: options.DrainTimeout, logger: options.Logger, metrics: options.Metrics, usage: options.UsageRecorder, affinity: options.Affinity, managedTunnels: options.ManagedTunnels, relayID: options.RelayID, affinityTTL: options.AffinityTTL, allowedOrigins: splitOrigins(options.AllowedOrigins), publicDomain: strings.TrimSuffix(strings.TrimSpace(options.PublicDomain), "."), bandwidth: engine.NewBandwidthLimiter(), orgLimits: make(map[string]int), orgConnections: make(map[string]int)}
	tcp.SetAdmissionHook(handler.allowConnection)
	tcp.SetUsageHook(func(tunnelID, eventType string, connections int) {
		organizationID, ok := router.OrganizationID(tunnelID)

		if ok && eventType == "tcp_connection_close" {
			handler.updateOrganizationConnections(organizationID, connections)
			handler.recordUsage(context.Background(), organizationID, tunnelID, eventType, 0, connections)
		}

	})
	return handler, nil
}

func (h *Handler) Metrics() *Metrics { return h.metrics }

func (h *Handler) Upgrade(c *fiber.Ctx) error {
	return websocket.New(h.Connect)(c)
}

func (h *Handler) Connect(connection *websocket.Conn) {

	if !h.originAllowed(connection.Headers("Origin")) {
		_ = connection.Close()
		return
	}

	if !h.acquireConnection() {
		_ = connection.WriteJSON(protocol.Envelope{Version: protocol.Version, Type: protocol.MessageTypeError, Payload: []byte(`{"code":"capacity","message":"relay capacity reached"}`)})
		_ = connection.Close()
		return
	}

	defer h.releaseConnection()
	h.metrics.AddConnection(1)
	defer h.metrics.AddConnection(-1)
	connection.SetReadLimit(h.maxFrameBytes)
	_ = connection.SetReadDeadline(time.Now().Add(h.readTimeout))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	identity := AgentIdentity{}

	if token := bearerToken(connection.Headers("Authorization")); token != "" {
		authenticatedIdentity, err := h.authenticator.Authenticate(ctx, token)

		if err != nil {
			h.writeJSON(connection, protocol.Envelope{Version: protocol.Version, Type: protocol.MessageTypeError, Payload: []byte(`{"code":"unauthorized","message":"invalid agent token"}`)})
			_ = connection.Close()
			return
		}

		identity = authenticatedIdentity
		h.logger.Info("relay connection authenticated", slog.String("agent_id", identity.AgentID), slog.String("organization_id", identity.OrganizationID))
		h.setOrganizationLimit(identity.OrganizationID, identity.MaxConnections)
	}

	owned := make(map[string]string)
	state := &connectionState{authenticated: identity.OrganizationID != "", identity: identity}
	connectionCtx, cancelConnection := context.WithCancel(ctx)
	defer cancelConnection()
	heartbeatStarted := false

	if state.authenticated {
		go h.sendHeartbeats(connectionCtx, connection, identity.OrganizationID)
		heartbeatStarted = true
	}

	defer func() { h.closeOwnedSessions(ctx, state.identity, owned) }()

	for {
		messageType, data, err := connection.ReadMessage()

		if err != nil {
			return
		}

		_ = connection.SetReadDeadline(time.Now().Add(h.readTimeout))
		h.metrics.AddFrame(int64(len(data)))

		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}

		message, err := protocol.Decode(data)

		if err != nil {
			h.metrics.AddError()
			h.writeError(connection, "protocol", err.Error())
			continue
		}

		if err := h.handleMessage(ctx, connection, state.identity, message, owned, state); err != nil {
			h.metrics.AddError()
			h.writeError(connection, "message", err.Error())
		}

		if state.authenticated && !heartbeatStarted {
			h.setOrganizationLimit(state.identity.OrganizationID, state.identity.MaxConnections)
			go h.sendHeartbeats(connectionCtx, connection, state.identity.OrganizationID)
			heartbeatStarted = true
		}

		h.recordMessageUsage(ctx, state.identity.OrganizationID, message)
	}
}

func (h *Handler) closeOwnedSessions(ctx context.Context, identity AgentIdentity, owned map[string]string) {

	for tunnelID, sessionID := range owned {

		if h.sessions.Remove(tunnelID, sessionID) {

			if h.affinity != nil {
				_ = h.affinity.Release(ctx, tunnelID, h.relayID)
			}

			h.recordUsage(ctx, identity.OrganizationID, tunnelID, "tunnel_close", 0, 0)
			h.metrics.AddTunnel(-1)
			h.router.RemoveTunnel(tunnelID)
			h.tcp.CloseTunnel(tunnelID)
			h.udp.CloseTunnel(tunnelID)
		}
	}

	h.logger.Info("relay connection closed", slog.String("agent_id", identity.AgentID), slog.Int("tunnels", len(owned)))
}

func (h *Handler) writeError(connection *websocket.Conn, code, message string) {
	payload, err := protocol.EncodePayload(protocol.MessageTypeError, "", protocol.ErrorMessage{Code: code, Message: message})

	if err == nil {
		_ = h.writeMessage(connection, websocket.TextMessage, payload)
	}
}

func (h *Handler) writeJSON(connection *websocket.Conn, value any) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return connection.WriteJSON(value)
}

func (h *Handler) writeMessage(connection *websocket.Conn, messageType int, data []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return connection.WriteMessage(messageType, data)
}

func (h *Handler) CloseAll() {

	for _, session := range h.sessions.Snapshot() {
		h.sessions.Remove(session.TunnelID, session.ID)
		h.router.RemoveTunnel(session.TunnelID)
		h.tcp.CloseTunnel(session.TunnelID)
		h.udp.CloseTunnel(session.TunnelID)
	}
}
