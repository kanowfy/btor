package metainfo_test

import (
	"bytes"
	"crypto/rand"
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
			name:  "correctly parses single file metainfo",
			input: []byte("d8:announce28:example.tracker.com/announce4:infod6:lengthi10000e4:name8:test.txt12:piece lengthi5000e6:pieces40:1111111111111111111122222222222222222222ee"),
			output: &metainfo.Metainfo{
				Announce: "example.tracker.com/announce",
				Info: metainfo.Info{
					Length:      10000,
					Name:        "test.txt",
					PieceLength: 5000,
					Pieces:      "1111111111111111111122222222222222222222",
				},
				InfoHash: []byte{84, 239, 11, 8, 172, 68, 174, 25, 34, 176, 23, 187, 185, 190, 201, 204, 180, 99, 219, 216},
			},
			fails: false,
		},
		{
			name:  "correctly parses multifile metainfo",
			input: []byte("d8:announce28:example.tracker.com/announce4:infod5:filesld6:lengthi1000e4:pathl3:foo7:bar.pngeed6:lengthi3000e4:pathl3:foo7:baz.jpgeee4:name4:test12:piece lengthi5000e6:pieces40:1111111111111111111122222222222222222222ee"),
			output: &metainfo.Metainfo{
				Announce: "example.tracker.com/announce",
				Info: metainfo.Info{
					Length: 4000,
					Files: []metainfo.FileEntry{
						{
							Length: 1000,
							Path:   []string{"foo", "bar.png"},
						},
						{
							Length: 3000,
							Path:   []string{"foo", "baz.jpg"},
						},
					},
					Name:        "test",
					PieceLength: 5000,
					Pieces:      "1111111111111111111122222222222222222222",
				},
				InfoHash:  []byte{19, 85, 72, 65, 98, 91, 234, 87, 65, 183, 110, 221, 118, 79, 78, 161, 144, 30, 178, 227},
				Multifile: true,
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
		{
			name:   "error on missing announce field",
			input:  []byte("d4:infod6:lengthi1000e4:name8:test.txt12:piece lengthi5000e6:pieces40:1111111111111111111122222222222222222222ee"),
			output: nil,
			fails:  true,
		},
		{
			name:   "error on missing info field",
			input:  []byte("d8:announce28:example.tracker.com/announcee"),
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

func TestPieceHashes(t *testing.T) {
	t.Parallel()

	bs := make([]byte, 40)
	if _, err := rand.Read(bs); err != nil {
		t.Fatal(err)
	}

	mi := &metainfo.Metainfo{
		Info: metainfo.Info{
			Pieces: string(bs),
		},
	}

	want := [][]byte{bs[0:20], bs[20:40]}

	got := mi.PieceHashes()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
