package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func probeICMP(ctx context.Context, target string) error {
	ip, err := resolveProbeIP(ctx, target)
	if err != nil {
		return err
	}
	var requestType icmp.Type = ipv4.ICMPTypeEcho
	var responseType icmp.Type = ipv4.ICMPTypeEchoReply
	network, address, protocol := "ip4:icmp", "0.0.0.0", 1
	if ip.To4() == nil {
		network, address, protocol = "ip6:ipv6-icmp", "::", 58
		requestType, responseType = ipv6.ICMPTypeEchoRequest, ipv6.ICMPTypeEchoReply
	}
	connection, err := icmp.ListenPacket(network, address)
	if err != nil {
		return fmt.Errorf("open ICMP socket (raw socket capability required): %w", err)
	}
	defer connection.Close()
	stop := context.AfterFunc(ctx, func() { connection.Close() })
	defer stop()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Second)
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set ICMP deadline: %w", err)
	}
	data := make([]byte, 24)
	if _, err := rand.Read(data); err != nil {
		return fmt.Errorf("generate ICMP challenge: %w", err)
	}
	id := int(data[0])<<8 | int(data[1])
	message := icmp.Message{Type: requestType, Code: 0, Body: &icmp.Echo{ID: id, Seq: 1, Data: data}}
	wire, err := message.Marshal(nil)
	if err != nil {
		return fmt.Errorf("encode ICMP echo: %w", err)
	}
	if _, err := connection.WriteTo(wire, &net.IPAddr{IP: ip}); err != nil {
		return fmt.Errorf("send ICMP echo: %w", err)
	}
	buffer := make([]byte, 4096)
	for {
		n, peer, err := connection.ReadFrom(buffer)
		if err != nil {
			return fmt.Errorf("receive ICMP echo: %w", err)
		}
		remote, ok := peer.(*net.IPAddr)
		if !ok || !remote.IP.Equal(ip) {
			continue
		}
		reply, err := icmp.ParseMessage(protocol, buffer[:n])
		if err != nil {
			continue
		}
		echo, ok := reply.Body.(*icmp.Echo)
		if reply.Type == responseType && ok && echo.ID == id && echo.Seq == 1 && bytes.Equal(echo.Data, data) {
			return nil
		}
	}
}
