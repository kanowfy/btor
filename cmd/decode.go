package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
	"github.com/spf13/cobra"
)

func decodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode [bencoded string]",
		Short: "decodes the provided bencoded string to stdout",
		Long:  "decodes the bencoded string and print the json encoded result to the standard out",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bencodeReader := strings.NewReader(args[0])

			var decoded interface{}
			if err := bencode.Unmarshal(bencodeReader, &decoded); err != nil {
				fmt.Printf("invalid bencoded string: %q\n", args[0])
				os.Exit(1)
			}

			jsonOutput, _ := json.Marshal(decoded)
			fmt.Println(string(jsonOutput))
		},
	}
}
