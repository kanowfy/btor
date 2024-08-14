package client

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/kanowfy/btor/handshake"
	"github.com/kanowfy/btor/message"
	"github.com/kanowfy/btor/peers"
)

const MaxBlockLen = 16384 // 2^14

// Client holds a connection with a peer
type Client struct {
	conn     net.Conn
	peer     peers.Peer
	infoHash []byte
	peerID   []byte
}

type PieceTask struct {
	Index  int
	Hash   []byte
	Length int
}

// New establish tcp connection with a peer and complete the handshake
func New(peer peers.Peer, infoHash, peerID []byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = handshake.InitHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		conn,
		peer,
		infoHash,
		peerID,
	}, nil
}

func (c *Client) DownloadPiece(pt PieceTask) ([]byte, error) {
	log.Println("waiting for bitfield message")
	// read bitfield
	if err := c.readBitfield(); err != nil {
		return nil, err
	}

	log.Println("sending interested message")
	// send interested
	if err := c.sendInterested(); err != nil {
		return nil, err
	}

	log.Println("waiting for unchoke message")
	// read unchoke
	if err := c.readUnchoke(); err != nil {
		return nil, err
	}

	log.Println("sending request messages")
	// send requests
	if err := c.sendRequests(pt.Index, pt.Length); err != nil {
		return nil, err
	}

	log.Println("reading piece messages")
	// read piece
	piece, err := c.readPiece(pt.Index, pt.Length)
	if err != nil {
		return nil, err
	}

	// check hash
	if err = matchPieceHash(piece, pt.Hash); err != nil {
		return nil, fmt.Errorf("invalid piece, mismatch piece hash")
	}

	return piece, nil
}

func CalculatePieceLength(pieceIndex, maxPieceLen, fileLen int) int {
	if fileLen/(pieceIndex+1) >= maxPieceLen {
		return maxPieceLen
	}

	return fileLen % maxPieceLen
}

func (c *Client) readBitfield() error {
	for {
		msg, err := message.Read(c.conn)
		if err != nil {
			return err
		}

		// skipping keep-alive message
		if msg == nil {
			continue
		}

		if msg.ID == message.MessageBitfield {
			break
		}
	}

	return nil
}

func (c *Client) sendInterested() error {
	msg := message.New(message.MessageInterested, nil)

	_, err := c.conn.Write(msg.Serialize())
	return err
}

func (c *Client) readUnchoke() error {
	for {
		msg, err := message.Read(c.conn)
		if err != nil {
			return err
		}

		if msg == nil {
			continue
		}

		if msg.ID == message.MessageUnchoke {
			break
		}
	}
	return nil
}

func (c *Client) sendRequests(pieceIndex, pieceLength int) error {
	var lengthSent int

	for lengthSent < pieceLength {
		blockLen := MaxBlockLen
		if pieceLength-lengthSent < MaxBlockLen {
			blockLen = pieceLength - lengthSent
		}
		msg := message.NewRequest(pieceIndex, lengthSent, blockLen)
		_, err := c.conn.Write(msg.Serialize())
		if err != nil {
			return err
		}
		lengthSent += blockLen
	}

	return nil
}

func (c *Client) readPiece(pieceIndex, pieceLength int) ([]byte, error) {
	buf := make([]byte, pieceLength)
	var lengthRead int
	for {
		msg, err := message.Read(c.conn)
		if err != nil {
			return nil, err
		}

		if msg.ID != message.MessagePiece {
			continue
		}

		rlen, err := message.ParsePiece(msg, buf, pieceIndex)
		if err != nil {
			return nil, err
		}

		lengthRead += rlen

		if lengthRead >= pieceLength {
			return buf, nil
		}
	}
}

func matchPieceHash(piece []byte, hash []byte) error {
	var hashed = sha1.New()
	hashed.Write(piece)
	checksum := hashed.Sum(nil)

	if !bytes.Equal(checksum, hash) {
		return fmt.Errorf("mismatch piece hash, want %x, got %x", hash, checksum)
	}

	return nil
}
