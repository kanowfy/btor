package cmd

import (
	"fmt"
	"os"

	"github.com/kanowfy/btor/torrent"
	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [torrent file]",
		Short: "print the information of the provided torrent file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			t, err := torrent.ParseFromFile(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			infoHash, err := t.InfoHash()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Tracker URL: %s\n", t.Announce)
			fmt.Printf("File Length: %d\n", t.Info.Length)
			fmt.Printf("Info Hash: %x\n", infoHash)

			pieceHashes := t.PieceHashes()
			fmt.Println("Piece Hashes:")
			for _, h := range pieceHashes {
				fmt.Printf("%x\n", h)
			}
		},
	}
}
