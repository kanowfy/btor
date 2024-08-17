package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"

	"github.com/kanowfy/btor/client"
	"github.com/kanowfy/btor/metainfo"
	"github.com/kanowfy/btor/peers"
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

			if err = downloadPiece(outfile, torrentfile, pieceIndex, peerID[:]); err != nil {
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

func downloadPiece(outFile string, torrentFile string, pieceIndex int, peerID []byte) error {
	f, err := os.Open(torrentFile)
	if err != nil {
		return err
	}
	defer f.Close()
	mi, err := metainfo.Parse(f)
	if err != nil {
		return err
	}

	infoHash, err := mi.InfoHash()
	if err != nil {
		return err
	}

	peerList, err := peers.Fetch(mi.Announce, infoHash, mi.Info.Length, peerID)
	if err != nil {
		return err
	}

	pieceHashes := mi.PieceHashes()

	// test with peer 0, assuming every peer has all the work
	peer := peerList[0]

	c, err := client.New(peer, infoHash, peerID)
	if err != nil {
		return err
	}

	pieceTask := client.PieceTask{
		Index:  pieceIndex,
		Hash:   pieceHashes[pieceIndex],
		Length: client.CalculatePieceLength(pieceIndex, mi.Info.PieceLength, mi.Info.Length),
	}

	piece, err := c.DownloadPiece(pieceTask)

	// write to dest
	if err = os.WriteFile(outFile, piece, 0o660); err != nil {
		return err
	}

	return nil
}
