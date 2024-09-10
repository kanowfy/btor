package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

const AppName = "btor"

func Execute() {
	var root = &cobra.Command{
		Use: AppName,
	}

	root.AddCommand(decodeCmd(), infoCmd(), peersCmd(), handshakeCmd(), downloadFileCmd())

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setupLogger(); err != nil {
			return err
		}

		return nil
	}

	root.Example = `Download from example.torrent file in home directory and save to example.txt file in the Download directory

	$ btor download ~/examplefile.torrent -o ~/Downloads/example.txt`
	root.Execute()
}

func setupLogger() error {
	logPath := getLogPath()

	if err := os.MkdirAll(filepath.Dir(logPath), os.ModePerm); err != nil {
		return err
	}

	out := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,
		MaxBackups: 2,
		MaxAge:     28,
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

func getLogPath() string {
	homeDir, _ := os.UserHomeDir()
	logfile := fmt.Sprintf("%s.log", AppName)

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(homeDir, ".local", "share", AppName, logfile)
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", AppName, logfile)
	case "darwin":
		return filepath.Join(homeDir, "Library", "Logs", AppName, logfile)
	default:
		return filepath.Join(homeDir, AppName, logfile)
	}
}
