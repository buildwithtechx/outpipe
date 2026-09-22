package engine

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type TCPConnections struct {
	mu          sync.RWMutex
	connections map[string]net.Conn
}

func NewTCPConnections() *TCPConnections {
	return &TCPConnections{connections: make(map[string]net.Conn)}
}

func (c *TCPConnections) Add(id string, connection net.Conn) error {

	if id == "" || connection == nil {
		return fmt.Errorf("connection id and connection are required")
	}

	c.mu.Lock()

	if _, exists := c.connections[id]; exists {
		c.mu.Unlock()
		return fmt.Errorf("connection %q already exists", id)
	}

	c.connections[id] = connection
	c.mu.Unlock()
	return nil
}

func (c *TCPConnections) Write(id string, data []byte) error {
	c.mu.RLock()
	connection, exists := c.connections[id]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection %q not found", id)
	}

	if _, err := connection.Write(data); err != nil {
		return fmt.Errorf("write tcp connection: %w", err)
	}

	return nil
}

func (c *TCPConnections) Remove(id string) error {
	c.mu.Lock()
	connection, exists := c.connections[id]

	if exists {
		delete(c.connections, id)
	}

	c.mu.Unlock()

	if !exists {
		return nil
	}

	return connection.Close()
}

type UDPPacket struct {
	ID        string
	Address   *net.UDPAddr
	ExpiresAt time.Time
}

type UDPPackets struct {
	mu      sync.Mutex
	packets map[string]UDPPacket
}

func NewUDPPackets() *UDPPackets {
	return &UDPPackets{packets: make(map[string]UDPPacket)}
}

func (p *UDPPackets) Add(packet UDPPacket) error {

	if packet.ID == "" || packet.Address == nil {
		return fmt.Errorf("udp packet id and address are required")
	}

	p.mu.Lock()
	p.packets[packet.ID] = packet
	p.mu.Unlock()
	return nil
}

func (p *UDPPackets) Take(id string, now time.Time) (UDPPacket, bool) {
	p.mu.Lock()
	packet, exists := p.packets[id]

	if exists {
		delete(p.packets, id)
	}

	p.mu.Unlock()

	if !exists || !packet.ExpiresAt.After(now) {
		return UDPPacket{}, false
	}

	return packet, true
}

func (p *UDPPackets) Cleanup(now time.Time) {
	p.mu.Lock()

	for id, packet := range p.packets {

		if !packet.ExpiresAt.After(now) {
			delete(p.packets, id)
		}
	}

	p.mu.Unlock()
}
