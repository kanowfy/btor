package bencode_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kanowfy/btor/bencode"
)

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
