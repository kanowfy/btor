package cmd

import "github.com/spf13/cobra"

func Execute() {
	var root = &cobra.Command{
		Use: "btor",
	}

	root.AddCommand(decodeCmd(), infoCmd(), peersCmd(), handshakeCmd())
	root.Execute()
}
