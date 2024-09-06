package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"github.com/kanowfy/btor/metainfo"
	"github.com/kanowfy/btor/peers"
	"github.com/spf13/cobra"
)

func peersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peers [torrent file]",
		Short: "fetch peers from tracker url and print to stard out",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			m, err := metainfo.Parse(f)
			if err != nil {
				if errors.Is(err, metainfo.ErrUnsupportedProtocol) {
					fmt.Println("protocol not supported")
				} else {
					fmt.Printf("could not read metainfo file: %v\n", err)
				}
				os.Exit(1)
			}

			var peerID [20]byte
			if _, err = rand.Read(peerID[:]); err != nil {
				panic(err)
			}

			peerList, err := peers.Fetch(m.Announce, m.InfoHash, m.Info.Length, peerID[:])
			if err != nil {
				fmt.Printf("failed to fetch peers: %v\n", err)
				os.Exit(1)
			}

			for _, p := range peerList {
				fmt.Printf("%s:%d\n", p.IP, p.Port)
			}

			return nil
		},
	}
}
