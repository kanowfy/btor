package cmd

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"

	"github.com/kanowfy/btor/client"
	"github.com/kanowfy/btor/metainfo"
	"github.com/kanowfy/btor/peers"
	"github.com/schollz/progressbar/v3"
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

	logger := slog.Default().With(slog.Group(
		"metainfo", slog.String("file_name", mi.Info.Name), slog.Int("file_size", mi.Info.Length),
	))

	taskStream := make(chan client.PieceTask, len(pieceHashes)) // put buffer to unblock
	resultStream := make(chan client.PieceResult)
	for _, peer := range peerList {
		go client.AttemptDownload(logger, peer, infoHash, peerID, taskStream, resultStream)
	}

	for i := 0; i < len(pieceHashes); i++ {
		taskStream <- client.PieceTask{
			Index:  i,
			Hash:   pieceHashes[i],
			Length: client.CalculatePieceLength(i, mi.Info.PieceLength, mi.Info.Length),
		}
	}

	resultBuf := make([]byte, mi.Info.Length)
	bar := progressbar.DefaultBytes(int64(mi.Info.Length), "downloading")

	var numResult int
	for numResult < len(pieceHashes) {
		res := <-resultStream
		numResult++

		start := mi.Info.PieceLength * res.Index
		end := start + res.Length
		copy(resultBuf[start:end], res.Data)
		bar.Write(res.Data)
	}
	close(taskStream)

	// write to dest
	if err = os.WriteFile(outFile, resultBuf, 0o660); err != nil {
		return err
	}

	return nil
}
