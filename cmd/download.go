package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"

	"github.com/kanowfy/btor/client"
	"github.com/spf13/cobra"
)

func downloadPieceCmd() *cobra.Command {
	var outfile string
	cmd := &cobra.Command{
		Use:   "download_piece -o OUT_FILE TORRENT_FILE PIECE_INDEX",
		Short: "download and save a piece",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			torrentfile := args[0]

			pieceIndex, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Printf("invalid piece index: %s", args[0])
				os.Exit(1)
			}

			var peerID [20]byte
			if _, err = rand.Read(peerID[:]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err = client.DownloadOne(outfile, torrentfile, pieceIndex, peerID[:]); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Piece %d downloaded to %s\n", pieceIndex, outfile)
		},
	}

	cmd.Flags().StringVarP(&outfile, "out", "o", "", "output file name")
	cmd.MarkFlagRequired("out")

	return cmd
}
