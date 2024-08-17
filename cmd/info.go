package cmd

import (
	"fmt"
	"os"

	"github.com/kanowfy/btor/metainfo"
	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [torrent file]",
		Short: "print the information of the provided torrent file",
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

			infoHash, err := m.InfoHash()
			if err != nil {
				return err
			}

			fmt.Printf("Tracker URL: %s\n", m.Announce)
			fmt.Printf("File Length: %d\n", m.Info.Length)
			fmt.Printf("Info Hash: %x\n", infoHash)

			pieceHashes := m.PieceHashes()
			fmt.Println("Piece Hashes:")
			for _, h := range pieceHashes {
				fmt.Printf("%x\n", h)
			}

			return nil
		},
	}
}
