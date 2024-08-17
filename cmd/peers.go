package cmd

import (
	"crypto/rand"
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
				return err
			}

			// calculate info hash
			infoHash, err := m.InfoHash()
			if err != nil {
				return err
			}

			var peerID [20]byte
			if _, err = rand.Read(peerID[:]); err != nil {
				return err
			}

			peerList, err := peers.Fetch(m.Announce, infoHash, m.Info.Length, peerID[:])
			if err != nil {
				return err
			}

			for _, p := range peerList {
				fmt.Printf("%s:%d\n", p.IP, p.Port)
			}

			return nil
		},
	}
}
