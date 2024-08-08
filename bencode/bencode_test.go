package bencode_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kanowfy/btor/bencode"
)

func TestDecode_CorrectlyDecodesBencodeString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		bencode string
		want    interface{}
	}{
		{
			name:    "integer",
			bencode: "i42e",
			want:    42,
		},
		{
			name:    "negative integer",
			bencode: "i-42e",
			want:    -42,
		},
		{
			name:    "string",
			bencode: "5:hello",
			want:    "hello",
		},
		{
			name:    "list",
			bencode: "l5:helloi42ee",
			want: []interface{}{
				"hello",
				42,
			},
		},
		{
			name:    "dictionary",
			bencode: "d5:hello5:world3:foo3:bare",
			want: map[string]interface{}{
				"hello": "world",
				"foo":   "bar",
			},
		},
		{
			name:    "nested dictionary",
			bencode: "d3:food3:bar3:baz6:numberi42eee",
			want: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar":    "baz",
					"number": 42,
				},
			},
		},
		{
			name:    "list of dictionary",
			bencode: "ld3:foo3:bared3:bar3:baz6:numberi42eee",
			want: []interface{}{
				map[string]interface{}{
					"foo": "bar",
				},
				map[string]interface{}{
					"bar":    "baz",
					"number": 42,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bencode.Unmarshal(tc.bencode)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Error(cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestEncode_CorrectlyEncodesDifferentTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "integer",
			input: 42,
			want:  "i42e",
		},
		{
			name:  "negative integer",
			input: -42,
			want:  "i-42e",
		},
		{
			name:  "string",
			input: "hello",
			want:  "5:hello",
		},
		{
			name: "list",
			input: []interface{}{
				"hello",
				42,
			},
			want: "l5:helloi42ee",
		},
		{
			name: "dictionary",
			input: map[string]interface{}{
				"hello": "world",
				"foo":   "bar",
			},
			want: "d3:foo3:bar5:hello5:worlde",
		},
		{
			name: "nested dictionary",
			input: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar":    "baz",
					"number": 42,
				},
			},
			want: "d3:food3:bar3:baz6:numberi42eee",
		},
		{
			name: "list of dictionary",
			input: []interface{}{
				map[string]interface{}{
					"foo": "bar",
				},
				map[string]interface{}{
					"bar":    "baz",
					"number": 42,
				},
			},
			want: "ld3:foo3:bared3:bar3:baz6:numberi42eee",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bencode.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, string(got)) {
				t.Error(cmp.Diff(tc.want, string(got)))
			}
		})
	}
}

func TestEncode_CorrectlyEncodeBackToOriginal(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../testdata/sample.torrent")
	if err != nil {
		t.Fatal(err)
	}

	res, err := bencode.Unmarshal(string(data))
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := bencode.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(data, encoded) {
		t.Error(data, encoded)
	}
}
