package cmd

import (
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Execute() {
	var root = &cobra.Command{
		Use: "btor",
	}

	root.AddCommand(decodeCmd(), infoCmd(), peersCmd(), handshakeCmd(), downloadFileCmd())

	var verbose bool
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setupLogger(verbose); err != nil {
			return err
		}

		return nil
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "whether to send logs to stdout")

	root.Example = `Download from example.torrent file in home directory and save to example.txt file in the Download directory

	$ btor download ~/examplefile.torrent -o ~/Downloads/example.txt`
	root.Execute()
}

func setupLogger(verbose bool) error {
	var out io.Writer

	fileout := &lumberjack.Logger{
		Filename:   os.ExpandEnv("$HOME/.local/share/btor/btor.log"),
		MaxSize:    100,
		MaxBackups: 2,
		MaxAge:     28,
	}

	if verbose {
		out = io.MultiWriter(os.Stdout, fileout)
	} else {
		out = fileout
	}

	logger := slog.New(
		slog.NewTextHandler(
			out,
			nil,
		),
	)

	slog.SetDefault(logger)
	return nil
}
