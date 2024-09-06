package peers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kanowfy/btor/peers"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		response []byte
		output   []peers.Peer
		fails    bool
	}{
		{
			name:     "correctly fetch and decode peers with regular presentation",
			response: []byte("d8:intervali5e5:peersld7:peer id20:123456789012345678902:ip7:1.2.3.44:porti12345eed7:peer id20:001122334455667788992:ip7:5.6.7.84:porti56789eeee"),
			output: []peers.Peer{
				{
					ID:   "12345678901234567890",
					IP:   "1.2.3.4",
					Port: 12345,
				},
				{
					ID:   "00112233445566778899",
					IP:   "5.6.7.8",
					Port: 56789,
				},
			},
			fails: false,
		},
		{
			name:     "correctly fetch and decode peers with compact representation",
			response: []byte{100, 56, 58, 105, 110, 116, 101, 114, 118, 97, 108, 105, 53, 101, 53, 58, 112, 101, 101, 114, 115, 49, 50, 58, 1, 2, 3, 4, 48, 57, 5, 6, 7, 8, 221, 213, 101},
			output: []peers.Peer{
				{
					IP:   "1.2.3.4",
					Port: 12345,
				},
				{
					IP:   "5.6.7.8",
					Port: 56789,
				},
			},
			fails: false,
		},
		{
			name:     "error on invalid bencode response",
			response: []byte("d8:intervali5e5:peersld7:peer id20:123456789012345678902:ip7:1.2.3.44:porti12345eed7:peer id20:001122334455667788992:ip7:5.6.7.84:porti56789"),
			output:   nil,
			fails:    true,
		},
		{
			name:     "error on unexpected response",
			response: []byte("d8:intervali5e5:peersi2ee"),
			output:   nil,
			fails:    true,
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		idx, _ := strconv.Atoi(path)

		w.Write(cases[idx].response)
	}))
	defer srv.Close()

	infoHash := []byte{214, 159, 145, 230, 178, 174, 76, 84, 36, 104, 209, 7, 58, 113, 212, 234, 19, 135, 154, 127}
	peerID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			peers, err := peers.Fetch(fmt.Sprintf("%s/%d", srv.URL, i), infoHash, 1000, peerID)
			if tc.fails {
				if err == nil {
					t.Fatal("expect error, got nil")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if !cmp.Equal(peers, tc.output) {
					t.Error(cmp.Diff(peers, tc.output))
				}
			}

		})
	}

}
