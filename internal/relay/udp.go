package relay

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/pkg/protocol"
)

type UDPManager struct {
	mu        sync.Mutex
	listeners map[string]*net.UDPConn
	packets   map[string]udpPacket
	senders   map[string]func(context.Context, protocol.Envelope) error
	max       int
}

type udpPacket struct {
	address  *net.UDPAddr
	listener *net.UDPConn
}

func NewUDPManager() *UDPManager {
	return &UDPManager{listeners: make(map[string]*net.UDPConn), packets: make(map[string]udpPacket), senders: make(map[string]func(context.Context, protocol.Envelope) error), max: 1000}
}

func (m *UDPManager) SetMaxPackets(max int) {

	if max < 1 {
		return
	}

	m.mu.Lock()
	m.max = max
	m.mu.Unlock()
}

func (m *UDPManager) Open(tunnelID string, send func(context.Context, protocol.Envelope) error) (int, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero})

	if err != nil {
		return 0, fmt.Errorf("listen public udp port: %w", err)
	}

	m.mu.Lock()
	m.listeners[tunnelID] = listener
	m.senders[tunnelID] = send
	m.mu.Unlock()
	go m.read(tunnelID, listener)
	return listener.LocalAddr().(*net.UDPAddr).Port, nil
}

func (m *UDPManager) read(tunnelID string, listener *net.UDPConn) {
	buffer := make([]byte, 64*1024)

	for {
		count, address, err := listener.ReadFromUDP(buffer)

		if err != nil {
			return
		}

		packetID := uuid.NewString()
		m.mu.Lock()

		if len(m.packets) >= m.max {
			m.mu.Unlock()
			continue
		}

		m.packets[packetID] = udpPacket{address: address, listener: listener}
		m.mu.Unlock()
		payload, encodeErr := protocol.EncodePayload(protocol.MessageTypeUDPData, "", protocol.UDPData{TunnelID: tunnelID, PacketID: packetID, SourceAddress: address.IP.String(), SourcePort: address.Port, Data: base64.StdEncoding.EncodeToString(buffer[:count])})
		send := m.sender(tunnelID)
		outgoing, decodeErr := protocol.Decode(payload)

		if encodeErr != nil || decodeErr != nil || send == nil || send(context.Background(), outgoing) != nil {
			m.mu.Lock()
			delete(m.packets, packetID)
			m.mu.Unlock()
			return
		}
	}
}

func (m *UDPManager) SetSender(tunnelID string, send func(context.Context, protocol.Envelope) error) {
	m.mu.Lock()

	if _, ok := m.listeners[tunnelID]; ok {
		m.senders[tunnelID] = send
	}

	m.mu.Unlock()
}

func (m *UDPManager) Port(tunnelID string) int {
	m.mu.Lock()
	listener := m.listeners[tunnelID]
	m.mu.Unlock()

	if listener == nil {
		return 0
	}

	return listener.LocalAddr().(*net.UDPAddr).Port
}

func (m *UDPManager) sender(tunnelID string) func(context.Context, protocol.Envelope) error {
	m.mu.Lock()
	send := m.senders[tunnelID]
	m.mu.Unlock()
	return send
}

func (m *UDPManager) Write(tunnelID string, response protocol.UDPResponse) error {

	if response.TunnelID != tunnelID {
		return fmt.Errorf("udp packet tunnel mismatch")
	}

	data, err := base64.StdEncoding.DecodeString(response.Data)

	if err != nil {
		return fmt.Errorf("decode udp data: %w", err)
	}

	m.mu.Lock()
	packet := m.packets[response.PacketID]
	delete(m.packets, response.PacketID)
	m.mu.Unlock()

	if packet.address == nil || packet.listener == nil {
		return fmt.Errorf("udp packet %q not found", response.PacketID)
	}

	_, err = packet.listener.WriteToUDP(data, packet.address)
	return err
}

func (m *UDPManager) CloseTunnel(tunnelID string) {
	m.mu.Lock()
	listener := m.listeners[tunnelID]
	delete(m.listeners, tunnelID)
	delete(m.senders, tunnelID)

	if listener != nil {

		for packetID, packet := range m.packets {

			if packet.listener == listener {
				delete(m.packets, packetID)
			}
		}
	}

	m.mu.Unlock()

	if listener != nil {
		_ = listener.Close()
	}
}
