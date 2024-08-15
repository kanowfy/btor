package handshake_test

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kanowfy/btor/handshake"
	"golang.org/x/sync/errgroup"
)

func TestNew(t *testing.T) {
	t.Parallel()

	infoHash := []byte{214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127}
	peerID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	want := &handshake.Handshake{
		Protocol: "BitTorrent protocol",
		Reserved: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	got := handshake.New(infoHash, peerID)

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSerialize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input *handshake.Handshake
		want  []byte
	}{
		{
			name: "standard protocol",
			input: &handshake.Handshake{
				Protocol: "BitTorrent protocol",
				Reserved: []byte{0, 0, 0, 0, 0, 0, 0, 0},
				InfoHash: []byte{214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127},
				PeerID:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			want: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			name: "weird protocol",
			input: &handshake.Handshake{
				Protocol: "New shiny BitTorrent protocol",
				Reserved: []byte{0, 0, 0, 0, 0, 0, 0, 0},
				InfoHash: []byte{214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127},
				PeerID:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},

			want: []byte{29, 78, 101, 119, 32, 115, 104, 105, 110, 121, 32, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Serialize()

			if !cmp.Equal(tc.want, got) {
				t.Error(cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestInitHandshake(t *testing.T) {
	t.Parallel()

	infoHash := []byte{214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127}
	peerIDSent := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peerIDRepl := []byte{0, 11, 22, 33, 44, 55, 66, 77, 88, 99, 0, 11, 22, 33, 44, 55, 66, 77, 88, 99}

	peerResponse := &handshake.Handshake{
		Protocol: "BitTorrent protocol",
		Reserved: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		InfoHash: infoHash,
		PeerID:   peerIDRepl,
	}

	cases := []struct {
		name   string
		reply  []byte
		output *handshake.Handshake
		fails  bool
	}{
		{
			name:   "correctly receives and parses handshake",
			reply:  peerResponse.Serialize(),
			output: peerResponse,
			fails:  false,
		},
		{
			name:   "error on empty reply",
			reply:  nil,
			output: nil,
			fails:  true,
		},
		{
			name:   "error on invalid reply",
			reply:  []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			output: nil,
			fails:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			self, peer := net.Pipe()

			var egr errgroup.Group

			egr.Go(func() error {
				defer peer.Close()
				buf := make([]byte, 1024)
				if _, err := peer.Read(buf); err != nil {
					return err
				}

				_, err := peer.Write(tc.reply)
				return err
			})

			defer self.Close()
			got, err := handshake.InitHandshake(self, infoHash, peerIDSent)
			if tc.fails {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if !cmp.Equal(tc.output, got) {
					t.Error(cmp.Diff(tc.output, got))
				}

			}

			if err := egr.Wait(); err != nil {
				t.Fatalf("peer error: %v", err)
			}

		})
	}

}
