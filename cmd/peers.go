package cmd

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/kanowfy/btor/peers"
	"github.com/kanowfy/btor/torrent"
	"github.com/spf13/cobra"
)

func peersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peers [torrent file]",
		Short: "fetch peers from tracker url and print to stard out",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			t, err := torrent.ParseFromFile(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// calculate info hash
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

			peerList, err := peers.Fetch(t.Announce, infoHash, t.Info.Length, peerID[:])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, p := range peerList {
				fmt.Printf("%s:%d\n", p.IP, p.Port)
			}
		},
	}
}
