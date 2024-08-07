package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/kanowfy/btor/torrent"
	"github.com/spf13/cobra"
)

func handshake() *cobra.Command {
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

			addr, err := parsePeerAddr(args[1])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			infoHash, err := t.InfoHash()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			reply, err := torrent.InitHandshake(addr, infoHash)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Peer ID: %x\n", reply.PeerID)
		},
	}
}

func parsePeerAddr(addr string) (*net.TCPAddr, error) {
	resolvedAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid peer address: %q", addr)
	}

	return resolvedAddr, nil
}
