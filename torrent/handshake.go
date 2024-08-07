package torrent

import (
	"bytes"
	"fmt"
	"net"
)

type Handshake struct {
	ProtocolLen byte
	Protocol    []byte
	Reserved    []byte
	InfoHash    []byte
	PeerID      []byte
}

func NewHandshake(infoHash []byte, peerID []byte) *Handshake {
	return &Handshake{
		ProtocolLen: byte(19),
		Protocol:    []byte("BitTorrent protocol"),
		Reserved:    make([]byte, 8),
		InfoHash:    infoHash,
		PeerID:      peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	var buf bytes.Buffer
	buf.WriteByte(h.ProtocolLen)
	buf.Write(h.Protocol)
	buf.Write(h.Reserved)
	buf.Write(h.InfoHash)
	buf.Write(h.PeerID)

	return buf.Bytes()
}

func DeserializeHandshake(msg []byte) (*Handshake, error) {
	if len(msg) != 68 {
		return nil, fmt.Errorf("invalid handshake reply: %s", msg)
	}

	return NewHandshake(msg[28:48], msg[48:68]), nil
}

func InitHandshake(addr *net.TCPAddr, infoHash []byte) (*Handshake, error) {
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msg := NewHandshake(infoHash, []byte(ID))

	_, err = conn.Write(msg.Serialize())
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 68)

	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	reply, err := DeserializeHandshake(buf[:n])
	if err != nil {
		return nil, err
	}

	return reply, nil
}
