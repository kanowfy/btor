package torrent

import (
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/kanowfy/btor/bencode"
	"github.com/mitchellh/mapstructure"
)

type Torrent struct {
	Announce string `mapstructure:"announce"`
	Info     struct {
		Length      int    `mapstructure:"length"`
		Name        string `mapstructure:"name"`
		PieceLength int    `mapstructure:"piece length"`
		Pieces      string `mapstructure:"pieces"`
	} `mapstructure:"info"`
}

func ParseFromFile(filename string) (*Torrent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	decoded, _, err := bencode.DecodeDict(string(data), 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t := &Torrent{}

	if err = mapstructure.Decode(decoded, &t); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Torrent) InfoHash() ([]byte, error) {
	var m map[string]interface{}
	mapstructure.Decode(t.Info, &m)
	infoBytes, err := bencode.Marshal(m)
	if err != nil {
		return nil, err
	}

	var infoHash = sha1.New()
	infoHash.Write(infoBytes)

	return infoHash.Sum(nil), nil
}

type PieceHash struct {
	Index int
	Hash  []byte
}

func (t *Torrent) PieceHashes() []PieceHash {
	hashes := make([]PieceHash, len(t.Info.Pieces)/20)
	for i := 0; i < len(t.Info.Pieces); i++ {
		hashes = append(hashes, PieceHash{
			Index: i,
			Hash:  []byte(t.Info.Pieces[i*20 : (i+1)*20]),
		})
	}

	return hashes
}
