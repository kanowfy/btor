package message_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestSetPiece_CorrectlySetValidIndex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		bitfield message.Bitfield
		index    int
	}{
		{
			name:     "set index of first octal",
			bitfield: message.Bitfield{0b00000000, 0b01010100},
			index:    0,
		},
		{
			name:     "set index of second octal",
			bitfield: message.Bitfield{0b11111111, 0b00000000},
			index:    9,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.bitfield.SetPieceIndex(tc.index)
			has := tc.bitfield.HasPiece(tc.index)
			if !has {
				t.Error("expected to have index (true) but got false")
			}
		})
	}
}

func TestSetPiece_DoesNothingWithInvalidIndex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		bitfield message.Bitfield
		index    int
	}{
		{
			name:     "set negative index",
			bitfield: message.Bitfield{0b00000000, 0b01010100},
			index:    -1,
		},
		{
			name:     "set index outside of range",
			bitfield: message.Bitfield{0b11111111, 0b00000000},
			index:    20,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := make(message.Bitfield, len(tc.bitfield))
			copy(before, tc.bitfield)

			tc.bitfield.SetPieceIndex(tc.index)

			if !cmp.Equal(before, tc.bitfield) {
				t.Error("expected to be the same but got differences", cmp.Diff(before, tc.bitfield))
			}

		})
	}
}
