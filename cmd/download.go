package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strconv"

	"github.com/kanowfy/btor/client"
	"github.com/kanowfy/btor/metainfo"
	"github.com/kanowfy/btor/peers"
	"github.com/spf13/cobra"
)

func downloadFileCmd() *cobra.Command {
	var outfile string
	cmd := &cobra.Command{
		Use:   "download -o OUT_FILE TORRENT_FILE",
		Short: "download and save a file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			torrentfile := args[0]

			var peerID [20]byte
			if _, err := rand.Read(peerID[:]); err != nil {
				return err
			}

			if err := downloadFile(outfile, torrentfile, peerID[:]); err != nil {
				return err
			}

			fmt.Printf("Downloaded %s to %s\n", torrentfile, outfile)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outfile, "out", "o", "", "output file name")
	cmd.MarkFlagRequired("out")

	return cmd
}

func downloadFile(outFile, torrentFile string, peerID []byte) error {
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

	ctx := context.WithValue(context.Background(), "peer", fmt.Sprintf("%s:%d", peer.IP, peer.Port))

	c, err := client.New(ctx, peer, infoHash, peerID)
	if err != nil {
		return err
	}

	tasks := make([]client.PieceTask, len(pieceHashes))
	for i := 0; i < len(tasks); i++ {
		tasks[i] = client.PieceTask{
			Index:  i,
			Hash:   pieceHashes[i],
			Length: client.CalculatePieceLength(i, mi.Info.PieceLength, mi.Info.Length),
		}
	}

	res, err := c.DownloadFile(ctx, tasks)
	if err != nil {
		return err
	}

	// write to dest
	if err = os.WriteFile(outFile, res, 0o660); err != nil {
		return err
	}

	return nil
}

func downloadPieceCmd() *cobra.Command {
	var outfile string
	cmd := &cobra.Command{
		Use:   "download_piece -o OUT_FILE TORRENT_FILE PIECE_INDEX",
		Short: "download and save a piece",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			torrentfile := args[0]

			pieceIndex, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid piece index: %v", err)
			}

			var peerID [20]byte
			if _, err = rand.Read(peerID[:]); err != nil {
				return err
			}

			if err = downloadPiece(outfile, torrentfile, pieceIndex, peerID[:]); err != nil {
				return err
			}

			fmt.Printf("Piece %d downloaded to %s\n", pieceIndex, outfile)
			return nil
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

	ctx := context.WithValue(context.Background(), "peer", fmt.Sprintf("%s:%d", peer.IP, peer.Port))

	c, err := client.New(ctx, peer, infoHash, peerID)
	if err != nil {
		return err
	}

	pieceTask := client.PieceTask{
		Index:  pieceIndex,
		Hash:   pieceHashes[pieceIndex],
		Length: client.CalculatePieceLength(pieceIndex, mi.Info.PieceLength, mi.Info.Length),
	}

	piece, err := c.DownloadPiece(ctx, pieceTask)
	if err != nil {
		return err
	}

	// write to dest
	if err = os.WriteFile(outFile, piece, 0o660); err != nil {
		return err
	}

	return nil
}
