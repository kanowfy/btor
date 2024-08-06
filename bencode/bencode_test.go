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
			got, _, err := bencode.Decode(tc.bencode, 0)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Error(cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestEncode_CorrectlyEncodesMap(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"announce": "announce",
		"info": map[string]interface{}{
			"length": 900,
			"name":   "sample.txt",
		},
	}

	want := []byte("d8:announce8:announce4:infod6:lengthi900e4:name10:sample.txtee")
	got, err := bencode.EncodeDict(input)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestEncode_CorrectlyEncodeBackToOriginal(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../testdata/sample.torrent")
	if err != nil {
		t.Fatal(err)
	}

	dict, _, err := bencode.DecodeDict(string(data), 0)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := bencode.EncodeDict(dict)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(data, encoded) {
		t.Error(data, encoded)
	}
}
