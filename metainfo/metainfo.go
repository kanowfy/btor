package metainfo

import (
	"crypto/sha1"
	"errors"
	"io"

	"github.com/jackpal/bencode-go"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
)

type Info struct {
	Length      int         `mapstructure:"length"`
	Files       []FileEntry `mapstructure:"files"`
	Name        string      `mapstructure:"name"`
	PieceLength int         `mapstructure:"piece length"`
	Pieces      string      `mapstructure:"pieces"`
}

type FileEntry struct {
	Length int      `mapstructure:"length"`
	Path   []string `mapstructure:"path"`
}

type Metainfo struct {
	Announce  string `mapstructure:"announce"`
	Info      Info   `mapstructure:"info"`
	InfoHash  []byte
	Multifile bool
}

// Parse parses a stream into Metainfo
func Parse(r io.Reader) (*Metainfo, error) {
	decoded, err := bencode.Decode(r)
	if err != nil {
		return nil, err
	}

	mi := &Metainfo{}

	if err = mapstructure.Decode(decoded, &mi); err != nil {
		return nil, err
	}

	if len(mi.Announce) == 0 {
		return nil, ErrUnsupportedProtocol
	}

	// calculate infohash
	m := make(map[string]interface{})
	if err = mapstructure.Decode(decoded, &m); err != nil {
		return nil, err
	}

	infoHash := sha1.New()
	if err = bencode.Marshal(infoHash, m["info"]); err != nil {
		return nil, err
	}

	mi.InfoHash = infoHash.Sum(nil)
	if mi.Info.Files != nil {
		mi.Multifile = true
		for _, file := range mi.Info.Files {
			mi.Info.Length += file.Length
		}
	}

	return mi, nil
}

// PieceHashes returns a hash slice of pieces in a file
func (m *Metainfo) PieceHashes() [][]byte {
	numPiece := len(m.Info.Pieces) / 20
	hashes := make([][]byte, numPiece)
	for i := 0; i < numPiece; i++ {
		hashes[i] = []byte(m.Info.Pieces[i*20 : (i+1)*20])
	}

	return hashes
}
