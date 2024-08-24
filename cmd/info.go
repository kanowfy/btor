package cmd

import (
	"errors"
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
				if errors.Is(err, metainfo.ErrUnsupportedProtocol) {
					fmt.Println("protocol not supported")
				} else {
					fmt.Printf("could not read metainfo file: %v\n", err)
				}
				os.Exit(1)
			}

			infoHash, err := m.InfoHash()
			if err != nil {
				fmt.Printf("failed to read info hash for metainfo file: %v\n", err)
				os.Exit(1)
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
