package message_test

import (
	"bytes"
	"errors"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
	"github.com/kanowfy/btor/message"
)

func TestSerialize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  *message.Message
		output []byte
	}{
		{
			name:   "message with no payload",
			input:  message.New(message.MessageInterested, nil),
			output: []byte{0, 0, 0, 5, 2},
		},
		{
			name:   "message with payload",
			input:  message.New(message.MessageRequest, []byte{0, 0, 0, 100}),
			output: []byte{0, 0, 0, 8, 6, 0, 0, 0, 100},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Serialize()

			if !cmp.Equal(tc.output, got) {
				cmp.Diff(tc.output, got)
			}
		})
	}
}

func TestRead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  []byte
		output *message.Message
		fails  bool
	}{
		{
			name:  "valid message with no payload",
			input: []byte{0, 0, 0, 1, 2},
			output: &message.Message{
				ID:      message.MessageInterested,
				Payload: []byte{},
			},
			fails: false,
		},
		{
			name:  "valid message with payload",
			input: []byte{0, 0, 0, 5, 6, 0, 0, 0, 100},
			output: &message.Message{
				ID:      message.MessageRequest,
				Payload: []byte{0, 0, 0, 100},
			},
			fails: false,
		},
		{
			name:   "no error on keep-alive message",
			input:  []byte{0, 0, 0, 0},
			output: nil,
			fails:  false,
		},
		{
			name:   "error on invalid message ID",
			input:  []byte{0, 0, 0, 1, 10},
			output: nil,
			fails:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.input)

			msg, err := message.Read(r)
			if tc.fails {
				if err == nil {
					t.Error("expect error, got nil")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if !cmp.Equal(tc.output, msg) {
					t.Error(cmp.Diff(tc.output, msg))
				}
			}
		})
	}
}

func TestRead_ReturnsReadError(t *testing.T) {
	t.Parallel()

	expected := errors.New("read error")

	r := iotest.ErrReader(expected)
	_, err := message.Read(r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expected) {
		t.Errorf("unexpected error, want %v got %v", expected, err)
	}
}

func TestParsePiece(t *testing.T) {
	t.Parallel()

	type input struct {
		msg        *message.Message
		buf        []byte
		pieceIndex int
	}

	cases := []struct {
		name   string
		input  input
		output int
		fails  bool
	}{
		{
			name: "correctly parses piece msg",
			input: input{
				msg: &message.Message{
					ID: message.MessagePiece,
					Payload: []byte{
						0, 0, 0, 0, // index
						0, 0, 0, 0, // begin
						1, 2, 3, 4, // data
					},
				},
				buf:        make([]byte, 4),
				pieceIndex: 0,
			},
			output: 4,
			fails:  false,
		},
		{
			name: "error on invalid msg id",
			input: input{
				msg: &message.Message{
					ID:      message.MessageBitfield,
					Payload: []byte{},
				},
				buf:        []byte{},
				pieceIndex: 0,
			},
			output: 0,
			fails:  true,
		},
		{
			name: "error on mismatch piece index",
			input: input{
				msg: &message.Message{
					ID: message.MessagePiece,
					Payload: []byte{
						0, 0, 0, 1,
						0, 0, 0, 0,
						1, 2, 3, 4,
					},
				},
				buf:        []byte{},
				pieceIndex: 0,
			},
			output: 0,
			fails:  true,
		},
		{
			name: "error on out of bound data",
			input: input{
				msg: &message.Message{
					ID: message.MessagePiece,
					Payload: []byte{
						0, 0, 0, 0,
						0, 0, 0, 0,
						1, 2, 3, 4,
					},
				},
				buf:        make([]byte, 3),
				pieceIndex: 0,
			},
			output: 0,
			fails:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			length, err := message.ParsePiece(tc.input.msg, tc.input.buf, tc.input.pieceIndex)
			if tc.fails {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if length != tc.output {
					t.Errorf("want %d, got %d", tc.output, length)
				}
			}
		})
	}
}
