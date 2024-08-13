package peers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/kanowfy/btor/bencode"
	"github.com/mitchellh/mapstructure"
)

type Peer struct {
	ID   string `mapstructure:"peer id"`
	IP   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

type trackerResponse struct {
	Interval int    `mapstructure:"interval"`
	Peers    []Peer `mapstructure:"peers"`
}

func Fetch(trackerUrl string, infoHash []byte, length int, peerID []byte) ([]Peer, error) {
	req, err := http.NewRequest(http.MethodGet, trackerUrl, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	q.Add("info_hash", string(infoHash))
	q.Add("peer_id", string(peerID))
	q.Add("port", "6881")
	q.Add("uploaded", "0")
	q.Add("downloaded", "0")
	q.Add("left", strconv.Itoa(length))
	q.Add("compact", "0")

	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decoded, err := bencode.Unmarshal(string(body))
	if err != nil {
		return nil, err
	}

	var tr trackerResponse
	if err = mapstructure.Decode(decoded, &tr); err != nil {
		return nil, fmt.Errorf("invalid tracker response: %+v", decoded)
	}

	return tr.Peers, nil
}
