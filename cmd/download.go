package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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
		Short: "download and save file from a .torrent file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			torrentfile := args[0]

			var peerID [20]byte
			if _, err := rand.Read(peerID[:]); err != nil {
				panic(err)
			}

			err := downloadFile(outfile, torrentfile, peerID[:])
			if err != nil {
				if errors.Is(err, metainfo.ErrUnsupportedProtocol) {
					fmt.Println("protocol not supported")
				} else {
					fmt.Printf("failed to download: %v\n", err)
				}
				os.Exit(1)
			}

			fmt.Printf("\nDownloaded %s to %s\n", torrentfile, outfile)
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

	peerList, err := peers.Fetch(mi.Announce, mi.InfoHash, mi.Info.Length, peerID)
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
		go client.StartDownloadClient(logger, peer, mi.InfoHash, peerID, taskStream, resultStream)
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

	// keep reading from resultStream until enough pieces are collected
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
	if mi.Multifile {
		var lengthWritten int
		for _, file := range mi.Info.Files {
			dirpath := filepath.Join(append([]string{outFile}, file.Path[:len(file.Path)-1]...)...)
			if err := os.MkdirAll(dirpath, 0o755); err != nil {
				return err
			}

			path := filepath.Join(dirpath, file.Path[len(file.Path)-1])

			if err = os.WriteFile(path, resultBuf[lengthWritten:lengthWritten+file.Length], 0o660); err != nil {
				return err
			}

			lengthWritten += file.Length
		}

	} else {
		if err = os.WriteFile(outFile, resultBuf, 0o660); err != nil {
			return err
		}
	}

	return nil
}
