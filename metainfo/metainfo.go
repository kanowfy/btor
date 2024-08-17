package metainfo

import (
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/kanowfy/btor/bencode"
	"github.com/mitchellh/mapstructure"
)

type Metainfo struct {
	Announce string `mapstructure:"announce"`
	Info     struct {
		Length      int    `mapstructure:"length"`
		Name        string `mapstructure:"name"`
		PieceLength int    `mapstructure:"piece length"`
		Pieces      string `mapstructure:"pieces"`
	} `mapstructure:"info"`
}

func ParseFromFile(filename string) (*Metainfo, error) {
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

	m := &Metainfo{}

	if err = mapstructure.Decode(decoded, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Metainfo) InfoHash() ([]byte, error) {
	var mp map[string]interface{}
	if err := mapstructure.Decode(m.Info, &mp); err != nil {
		return nil, fmt.Errorf("error decoding info to map: %v", err)
	}
	infoBytes, err := bencode.Marshal(m)
	if err != nil {
		return nil, err
	}

	var infoHash = sha1.New()
	infoHash.Write(infoBytes)

	return infoHash.Sum(nil), nil
}

func (m *Metainfo) PieceHashes() [][]byte {
	numPiece := len(m.Info.Pieces) / 20
	hashes := make([][]byte, numPiece)
	for i := 0; i < numPiece; i++ {
		hashes[i] = []byte(m.Info.Pieces[i*20 : (i+1)*20])
	}

	return hashes
}
