package client

import (
	"encoding/base64"
	"net"

	"outpipe.dev/outpipe/pkg/protocol"
)

func (c *RelayConnection) handleUDPData(target string, message protocol.Envelope) error {
	var data protocol.UDPData
	if err := protocol.DecodePayload(message, &data); err != nil {
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(data.Data)
	if err != nil {
		return err
	}
	c.udpMu.Lock()
	created := c.udpConn == nil
	if created {
		address, resolveErr := net.ResolveUDPAddr("udp", target)
		if resolveErr != nil {
			c.udpMu.Unlock()
			return resolveErr
		}
		c.udpConn, err = net.DialUDP("udp", nil, address)
	}
	if err == nil {
		c.udpQueue = append(c.udpQueue, data.PacketID)
	}
	connection := c.udpConn
	c.udpMu.Unlock()
	if err != nil {
		return err
	}
	if created {
		go c.proxyUDP(connection)
	}
	_, err = connection.Write(decoded)
	return err
}

func (c *RelayConnection) proxyUDP(connection *net.UDPConn) {
	buffer := make([]byte, 64*1024)
	for {
		count, address, err := connection.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		c.udpMu.Lock()
		if len(c.udpQueue) == 0 {
			c.udpMu.Unlock()
			continue
		}
		packetID := c.udpQueue[0]
		c.udpQueue = c.udpQueue[1:]
		c.udpMu.Unlock()
		payload, encodeErr := protocol.EncodePayload(protocol.MessageTypeUDPResponse, "", protocol.UDPResponse{TunnelID: c.TunnelID, PacketID: packetID, TargetAddress: address.IP.String(), TargetPort: address.Port, Data: base64.StdEncoding.EncodeToString(buffer[:count])})
		if encodeErr != nil || c.write(payload) != nil {
			return
		}
	}
}

func (c *RelayConnection) handleTCPData(target string, message protocol.Envelope) error {
	var data protocol.TCPData
	if err := protocol.DecodePayload(message, &data); err != nil {
		return err
	}
	c.tcpMu.Lock()
	connection := c.tcpConns[data.ConnectionID]
	created := false
	var err error
	if connection == nil {
		connection, err = net.Dial("tcp", target)
		if err == nil {
			c.tcpConns[data.ConnectionID] = connection
			created = true
		}
	}
	c.tcpMu.Unlock()
	if connection == nil {
		return c.sendTCPClose(data.ConnectionID, "connect local target failed")
	}
	decoded, err := base64.StdEncoding.DecodeString(data.Data)
	if err != nil {
		c.closeTCP(data.ConnectionID)
		return err
	}
	if _, err := connection.Write(decoded); err != nil {
		c.closeTCP(data.ConnectionID)
		return err
	}
	if created {
		go c.proxyTCP(data.ConnectionID, connection)
	}
	return nil
}

func (c *RelayConnection) proxyTCP(connectionID string, connection net.Conn) {
	defer connection.Close()
	buffer := make([]byte, 32*1024)
	for {
		count, err := connection.Read(buffer)
		if count > 0 {
			payload, encodeErr := protocol.EncodePayload(protocol.MessageTypeTCPData, "", protocol.TCPData{TunnelID: c.TunnelID, ConnectionID: connectionID, Data: base64.StdEncoding.EncodeToString(buffer[:count])})
			if encodeErr != nil || c.write(payload) != nil {
				return
			}
		}
		if err != nil {
			c.closeTCP(connectionID)
			_ = c.sendTCPClose(connectionID, err.Error())
			return
		}
	}
}

func (c *RelayConnection) closeTCP(connectionID string) {
	c.tcpMu.Lock()
	connection := c.tcpConns[connectionID]
	delete(c.tcpConns, connectionID)
	c.tcpMu.Unlock()
	if connection != nil {
		_ = connection.Close()
	}
}

func (c *RelayConnection) sendTCPClose(connectionID, reason string) error {
	payload, err := protocol.EncodePayload(protocol.MessageTypeTCPClose, "", protocol.TCPClose{TunnelID: c.TunnelID, ConnectionID: connectionID, Reason: reason})
	if err != nil {
		return err
	}
	return c.write(payload)
}
