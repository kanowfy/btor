package message_test

import (
	"testing"

	"github.com/kanowfy/btor/message"
)

func TestHasPiece(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		bitfield message.Bitfield
		index    int
		want     bool
	}{
		{
			name:     "include with index 0",
			bitfield: message.Bitfield{0b10000000, 0b01010100},
			index:    0,
			want:     true,
		},
		{
			name:     "not include with index 0",
			bitfield: message.Bitfield{0b01111111, 0b01010100},
			index:    0,
			want:     false,
		},
		{
			name:     "include with index 9",
			bitfield: message.Bitfield{0b10000000, 0b01010100},
			index:    0,
			want:     true,
		},
		{
			name:     "not include with index 9",
			bitfield: message.Bitfield{0b01000000, 0b00010100},
			index:    0,
			want:     false,
		},
		{
			name:     "invalid index, negative",
			bitfield: message.Bitfield{0b01000000, 0b00010100},
			index:    -5,
			want:     false,
		},
		{
			name:     "invalid index, out of range",
			bitfield: message.Bitfield{0b01000000, 0b00010100},
			index:    16,
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.bitfield.HasPiece(tc.index)
			if got != tc.want {
				t.Errorf("expected to be %v, but got %v", tc.want, got)
			}
		})
	}
}
