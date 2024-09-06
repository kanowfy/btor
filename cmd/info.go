package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

			fmt.Printf("• Name: %s\n", m.Info.Name)
			fmt.Printf("• Tracker URL: %s\n", m.Announce)
			if !m.Multifile {
				fmt.Printf("• File Length: %d\n", m.Info.Length)
			} else {
				var size int
				for _, f := range m.Info.Files {
					size += f.Length
				}
				fmt.Printf("• Total Files/Size: %d files/%dB\n", len(m.Info.Files), size)
				fmt.Println("• Files:")
				for i := range len(m.Info.Files) {
					fmt.Printf("%d. %s/%s\n", i+1, m.Info.Name, strings.Join(m.Info.Files[i].Path, "/"))
				}
			}
			/*
				fmt.Printf("Info Hash: %x\n", m.InfoHash)

				pieceHashes := m.PieceHashes()
				fmt.Println("Piece Hashes:")
				for _, h := range pieceHashes {
					fmt.Printf("%x\n", h)
				}
			*/

			return nil
		},
	}
}
