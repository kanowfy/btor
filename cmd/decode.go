package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/kanowfy/btor/bencode"
	"github.com/spf13/cobra"
)

func decodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode [bencoded string]",
		Short: "decodes the provided bencoded string to stdout",
		Long:  "decodes the bencoded string and print the json encoded result to the standard out",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bencodedValue := args[0]

			decoded, err := bencode.Unmarshal(bencodedValue)
			if err != nil {
				return err
			}
			jsonOutput, _ := json.Marshal(decoded)
			fmt.Println(string(jsonOutput))
			return nil
		},
	}
}
