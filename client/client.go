package client

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/kanowfy/btor/handshake"
	"github.com/kanowfy/btor/message"
	"github.com/kanowfy/btor/peers"
	"github.com/kanowfy/btor/torrent"
)

const MaxBlockLen = 16384

func DownloadOne(outFile string, torrentFile string, pieceIndex int, peerID []byte) error {
	torrent, err := torrent.ParseFromFile(torrentFile)
	if err != nil {
		return err
	}

	infoHash, err := torrent.InfoHash()
	if err != nil {
		return err
	}

	peerList, err := peers.Fetch(torrent.Announce, infoHash, torrent.Info.Length, peerID)
	if err != nil {
		return err
	}

	pieceHashes := torrent.PieceHashes()

	// test with peer 0
	peer := peerList[0]

	pieceLen := pieceLengthByIndex(pieceIndex, torrent.Info.PieceLength, torrent.Info.Length)

	piece, err := downloadPiece(pieceIndex, pieceLen, pieceHashes[pieceIndex], peer, infoHash, peerID)

	// write to dest
	if err = os.WriteFile(outFile, piece, 0o660); err != nil {
		return err
	}

	return nil
}

func downloadPiece(pieceIndex, pieceLen int, pieceHash []byte, peer peers.Peer, infoHash, peerID []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	log.Println("initiating handshake")
	if _, err = handshake.InitHandshake(conn, infoHash, peerID); err != nil {
		return nil, err
	}

	log.Println("waiting for bitfield message")
	// read bitfield
	if err = readBitfield(conn); err != nil {
		return nil, err
	}

	log.Println("sending interested message")
	// send interested
	if err = sendInterested(conn); err != nil {
		return nil, err
	}

	log.Println("waiting for unchoke message")
	// read unchoke
	if err = readUnchoke(conn); err != nil {
		return nil, err
	}

	log.Println("sending request messages")
	// send requests
	if err = sendRequests(conn, pieceIndex, pieceLen); err != nil {
		return nil, err
	}

	log.Println("reading piece messages")
	// read piece
	piece, err := readPiece(conn, pieceIndex, pieceLen)
	if err != nil {
		return nil, err
	}

	// check hash
	if err = matchPieceHash(piece, pieceHash); err != nil {
		return nil, fmt.Errorf("invalid piece, mismatch piece hash")
	}

	return piece, nil
}

func pieceLengthByIndex(pieceIndex, maxPieceLen, fileLen int) int {
	if fileLen/(pieceIndex+1) >= maxPieceLen {
		return maxPieceLen
	}

	return fileLen % maxPieceLen
}

func readBitfield(r io.Reader) error {
	for {
		msg, err := message.Read(r)
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

func sendInterested(w io.Writer) error {
	msg := message.New(message.MessageInterested, nil)

	_, err := w.Write(msg.Serialize())
	return err
}

func readUnchoke(r io.Reader) error {
	for {
		msg, err := message.Read(r)
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

func sendRequests(w io.Writer, pieceIndex int, pieceLength int) error {
	var lengthSent int

	for lengthSent < pieceLength {
		blockLen := MaxBlockLen
		if pieceLength-lengthSent < MaxBlockLen {
			blockLen = pieceLength - lengthSent
		}
		msg := message.NewRequest(pieceIndex, lengthSent, blockLen)
		_, err := w.Write(msg.Serialize())
		if err != nil {
			return err
		}
		lengthSent += blockLen
	}

	return nil
}

func readPiece(r io.Reader, pieceIndex int, pieceLength int) ([]byte, error) {
	buf := make([]byte, pieceLength)
	var lengthRead int
	for {
		msg, err := message.Read(r)
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
