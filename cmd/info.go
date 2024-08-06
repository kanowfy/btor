package cmd

import (
	"fmt"
	"os"

	"github.com/kanowfy/btor/torrent"
	"github.com/spf13/cobra"
)

func info() *cobra.Command {
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

			// calculate info hash
			infoHash, err := t.InfoHash()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Tracker URL: %s\n", t.Announce)
			fmt.Printf("File Length: %d\n", t.Info.Length)
			fmt.Printf("Info Hash: %x\n", infoHash)

			fmt.Println("Piece Hashes:")
			for i := 0; i < len(t.Info.Pieces)/20; i++ {
				fmt.Printf("%x\n", t.Info.Pieces[i*20:(i+1)*20])
			}
		},
	}
}
