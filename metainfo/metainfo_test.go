package metainfo_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kanowfy/btor/metainfo"
)

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  []byte
		output *metainfo.Metainfo
		fails  bool
	}{
		{
			name:  "correctly parses metainfo",
			input: []byte("d8:announce28:example.tracker.com/announce4:infod6:lengthi10000e4:name8:test.txt12:piece lengthi5000e6:pieces40:1111111111111111111122222222222222222222ee"),
			output: &metainfo.Metainfo{
				Announce: "example.tracker.com/announce",
				Info: metainfo.Info{
					Length:      10000,
					Name:        "test.txt",
					PieceLength: 5000,
					Pieces:      "1111111111111111111122222222222222222222",
				},
			},
			fails: false,
		},
		{
			name:   "error on invalid bencode",
			input:  []byte("foobar"),
			output: nil,
			fails:  true,
		},
		{
			name:   "error on invalid content",
			input:  []byte("d8:announce28:example.tracker.com/announce4:infod6:length3:foo4:name8:test.txt12:piece lengthi5000e6:pieces40:1111111111111111111122222222222222222222ee"), // length is string
			output: nil,
			fails:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewReader(tc.input)
			m, err := metainfo.Parse(reader)
			if tc.fails {
				if err == nil {
					t.Fatal("expect error, got nil")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if !cmp.Equal(tc.output, m) {
					t.Error(cmp.Diff(tc.output, m))
				}
			}
		})
	}
}
