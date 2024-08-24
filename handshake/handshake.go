package handshake

import (
	"bytes"
	"fmt"
	"io"
)

type Handshake struct {
	Protocol string
	Reserved []byte
	InfoHash []byte
	PeerID   []byte
}

func New(infoHash []byte, peerID []byte) *Handshake {
	return &Handshake{
		Protocol: "BitTorrent protocol",
		Reserved: make([]byte, 8),
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(len(h.Protocol)))
	buf.WriteString(h.Protocol)
	buf.Write(h.Reserved)
	buf.Write(h.InfoHash)
	buf.Write(h.PeerID)

	return buf.Bytes()
}

func deserialize(msg []byte) (*Handshake, error) {
	if len(msg) != 68 {
		return nil, fmt.Errorf("invalid handshake reply: %s", msg)
	}

	return New(msg[28:48], msg[48:68]), nil
}

// InitHandshake sends a handshake message to a stream, then reads and
// returns a handshake reply
func InitHandshake(rw io.ReadWriter, infoHash []byte, peerID []byte) (*Handshake, error) {
	msg := New(infoHash, peerID)

	_, err := rw.Write(msg.Serialize())
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 68)

	n, err := rw.Read(buf)
	if err != nil {
		return nil, err
	}

	reply, err := deserialize(buf[:n])
	if err != nil {
		return nil, err
	}

	return reply, nil
}
