package cmd

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"

	"github.com/kanowfy/btor/handshake"
	"github.com/kanowfy/btor/torrent"
	"github.com/spf13/cobra"
)

func handshakeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "handshake [torrent file] <peer_ip>:<peer_port>",
		Short: "send a handshake message to the peer destination and print the replied peer ID in hexadecimal",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			t, err := torrent.ParseFromFile(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			infoHash, err := t.InfoHash()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var peerID [20]byte
			if _, err = rand.Read(peerID[:]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			reply, err := getHandshakeMessage(args[1], infoHash, peerID[:])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Peer ID: %x\n", reply.PeerID)
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
