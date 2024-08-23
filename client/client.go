package client

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log/slog"
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
	bitfield message.Bitfield
	logger   *slog.Logger
}

type PieceTask struct {
	Index  int
	Hash   []byte
	Length int
}

// New establish tcp connection with a peer and complete the handshake
func New(logger *slog.Logger, peer peers.Peer, infoHash, peerID []byte) (*Client, error) {
	logger = logger.With(slog.String("peer_addr", fmt.Sprintf("%s:%d", peer.IP, peer.Port)))

	logger.Info("establishing connection with peer")
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port), 3*time.Second)
	if err != nil {
		logger.Error("failed to establish connection with peer", "error", err)
		return nil, err
	}

	logger.Info("performing handshake with peer")
	_, err = handshake.InitHandshake(conn, infoHash, peerID)
	if err != nil {
		logger.Error("failed to complete handshake with peer", "error", err)
		conn.Close()
		return nil, err
	}

	msg, err := readBitfield(conn)
	if err != nil {
		logger.Error("failed to read bitfield message from peer", "error", err)
		return nil, err
	}

	bitfield := message.Bitfield(msg.Payload)

	return &Client{
		conn,
		peer,
		infoHash,
		peerID,
		bitfield,
		logger,
	}, nil
}

func (c *Client) DownloadFile(tasks []PieceTask) ([]byte, error) {
	var resBuf bytes.Buffer
	// send interested
	if err := c.sendInterested(); err != nil {
		c.logger.Error("failed to send interested message to peer", "error", err)
		return nil, err
	}

	// read unchoke
	if err := c.readUnchoke(); err != nil {
		c.logger.Error("failed to read unchoke message from peer", "error", err)
		return nil, err
	}

	for _, pt := range tasks {
		c.logger.Info("requesting piece", slog.Int("piece index", pt.Index))
		// send requests
		if err := c.sendRequests(pt.Index, pt.Length); err != nil {
			c.logger.Error("failed to send requests message to peer", "error", err)
			return nil, err
		}

		c.logger.Info("downloading piece", slog.Int("piece index", pt.Index))
		// read piece
		piece, err := c.readPiece(pt.Index, pt.Length)
		if err != nil {
			c.logger.Error("failed to read piece message from peer", "error", err)
			return nil, err
		}
		// check hash
		if err := matchPieceHash(piece, pt.Hash); err != nil {
			c.logger.Error("failed to validate piece hash", "error", err)
			return nil, fmt.Errorf("invalid piece, mismatch piece hash")
		}

		if _, err := resBuf.Write(piece); err != nil {
			c.logger.Error("failed to write piece data to buffer", "error", err)
			return nil, err
		}
		c.logger.Info("piece downloaded", slog.Int("piece index", pt.Index))
	}

	return resBuf.Bytes(), nil
}

// DownloadPiece attempts to download a piece in a blocking manner
func (c *Client) DownloadPiece(pt PieceTask) ([]byte, error) {
	// send interested
	if err := c.sendInterested(); err != nil {
		c.logger.Error("failed to send interested message to peer", "error", err)
		return nil, err
	}

	// read unchoke
	if err := c.readUnchoke(); err != nil {
		c.logger.Error("failed to read unchoke message from peer", "error", err)
		return nil, err
	}

	// send requests
	if err := c.sendRequests(pt.Index, pt.Length); err != nil {
		c.logger.Error("failed to send requests message to peer", "error", err)
		return nil, err
	}

	// read piece
	piece, err := c.readPiece(pt.Index, pt.Length)
	if err != nil {
		c.logger.Error("failed to read piece message from peer", "error", err)
		return nil, err
	}

	// check hash
	if err = matchPieceHash(piece, pt.Hash); err != nil {
		c.logger.Error("failed to validate piece hash", "error", err)
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

func readBitfield(conn net.Conn) (*message.Message, error) {
	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("expected bitfield message but got nothing")
	}

	if msg.ID != message.MessageBitfield {
		return nil, fmt.Errorf("expected bitfield message but got message with ID %d", msg.ID)
	}

	return msg, nil
}

func (c *Client) sendInterested() error {
	msg := message.New(message.MessageInterested, nil)

	_, err := c.conn.Write(msg.Serialize())
	return err
}

func (c *Client) readUnchoke() error {
	msg, err := message.Read(c.conn)
	if err != nil {
		return err
	}

	if msg == nil {
		return fmt.Errorf("expected unchoke message but got nothing")
	}

	if msg.ID != message.MessageUnchoke {
		return fmt.Errorf("expected unchoke message but got message with ID %d", msg.ID)
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
