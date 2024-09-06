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

const (
	MaxBlockLen       = 16384 // 2^14
	MaxPipelinedBlock = 5
)

// Client holds a connection with a peer
type Client struct {
	conn         net.Conn
	peer         peers.Peer
	infoHash     []byte
	peerID       []byte
	bitfield     message.Bitfield
	recvBitfield bool
	choke        bool
	logger       *slog.Logger
}

type PieceTask struct {
	Index  int
	Hash   []byte
	Length int
}

type PieceResult struct {
	PieceTask
	Data []byte
}

type pieceState struct {
	index      int
	client     *Client
	requested  int
	downloaded int
	pipelined  int
	buf        []byte
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

	msg, err := readFirstMsg(conn)
	if err != nil {
		logger.Error("failed to read message from peer", "error", err)
		return nil, err
	}

	var bitfield message.Bitfield
	var recvBitfield bool
	if msg.ID == message.MessageBitfield {
		bitfield = message.Bitfield(msg.Payload)
		recvBitfield = true
	}

	return &Client{
		conn,
		peer,
		infoHash,
		peerID,
		bitfield,
		recvBitfield,
		false,
		logger,
	}, nil
}

func StartDownloadClient(logger *slog.Logger, peer peers.Peer, infoHash, peerID []byte, taskStream chan PieceTask, resultStream chan<- PieceResult) {
	c, err := New(logger, peer, infoHash, peerID)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	// send interested
	if err := c.sendInterested(); err != nil {
		c.logger.Error("failed to send interested message to peer", "error", err)
		return
	}

	// read unchoke if haven't
	if c.recvBitfield {
		if err := c.readUnchoke(); err != nil {
			c.logger.Error("failed to read unchoke message from peer", "error", err)
			return
		}
	}

	for pt := range taskStream {
		if c.recvBitfield {
			if !c.bitfield.HasPiece(pt.Index) {
				// put task back to queue
				taskStream <- pt
				continue
			}
		}

		c.logger.Info("downloading piece", slog.Int("index", pt.Index))
		piece, err := downloadPiece(c, pt)
		if err != nil {
			c.logger.Info("could not download piece", slog.Int("index", pt.Index), "error", err)
			taskStream <- pt
			return
		}

		// check hash
		if err := matchPieceHash(piece, pt.Hash); err != nil {
			c.logger.Error("failed to validate piece hash", "error", err)
			taskStream <- pt
			continue
		}

		c.logger.Info("piece downloaded", slog.Int("piece index", pt.Index))
		c.sendHave(pt.Index)
		resultStream <- PieceResult{
			PieceTask: pt,
			Data:      piece,
		}
	}
}

func downloadPiece(client *Client, pt PieceTask) ([]byte, error) {
	state := pieceState{
		index:  pt.Index,
		client: client,
		buf:    make([]byte, pt.Length),
	}

	for state.downloaded < pt.Length {
		if !state.client.choke {
			for state.pipelined < MaxPipelinedBlock && state.requested < pt.Length {

				blockLen := MaxBlockLen
				if pt.Length-state.requested < MaxBlockLen {
					blockLen = pt.Length - state.requested
				}

				if err := state.client.sendRequest(pt.Index, state.requested, blockLen); err != nil {
					return nil, err
				}
				state.requested += blockLen
				state.pipelined++
			}
		}

		if err := state.readMessage(); err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (state *pieceState) readMessage() error {
	msg, err := message.Read(state.client.conn)
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MessageChoke:
		state.client.choke = true
	case message.MessageUnchoke:
		state.client.choke = false
	case message.MessageHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}

		state.client.bitfield.SetPieceIndex(index)
		// to be implemented
	case message.MessagePiece:
		got, err := message.ParsePiece(msg, state.buf, state.index)
		if err != nil {
			return err
		}

		state.downloaded += got
		state.pipelined--
	}

	return nil
}

func CalculatePieceLength(pieceIndex, maxPieceLen, fileLen int) int {
	if fileLen/(pieceIndex+1) >= maxPieceLen {
		return maxPieceLen
	}

	return fileLen % maxPieceLen
}

func readFirstMsg(conn net.Conn) (*message.Message, error) {
	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("expected message but got nothing")
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

func (c *Client) sendRequest(pieceIndex, offset, pieceLength int) error {
	msg := message.NewRequest(pieceIndex, offset, pieceLength)
	_, err := c.conn.Write(msg.Serialize())
	return err
}

func (c *Client) sendHave(pieceIndex int) error {
	msg := message.NewHave(pieceIndex)
	_, err := c.conn.Write(msg.Serialize())
	return err
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
