package cmd

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"

	"github.com/kanowfy/btor/handshake"
	"github.com/kanowfy/btor/metainfo"
	"github.com/spf13/cobra"
)

func handshakeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "handshake [torrent file] <peer_ip>:<peer_port>",
		Short: "perform handshake with a peer and print out the received peer id",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()
			m, err := metainfo.Parse(f)
			if err != nil {
				return err
			}

			infoHash, err := m.InfoHash()
			if err != nil {
				return err
			}

			var peerID [20]byte
			if _, err = rand.Read(peerID[:]); err != nil {
				return err
			}

			reply, err := getHandshakeMessage(args[1], infoHash, peerID[:])
			if err != nil {
				return err
			}

			fmt.Printf("Peer ID: %x\n", reply.PeerID)
			return nil
		},
	}
}

func getHandshakeMessage(addr string, infoHash []byte, peerID []byte) (*handshake.Handshake, error) {
	resolvedAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid peer address: %q", addr)
	}

	conn, err := net.DialTCP("tcp", nil, resolvedAddr)
	if err != nil {
		return nil, err
	}

	h, err := handshake.InitHandshake(conn, infoHash, peerID)
	if err != nil {
		return nil, err
	}

	return h, nil
}
