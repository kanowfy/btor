package torrent

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/kanowfy/btor/bencode"
	"github.com/mitchellh/mapstructure"
)

// const ID = "10111213141516171819"
const ID = "00112233445566778899"

type Peer struct {
	ID   string `mapstructure:"peer id"`
	IP   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

type TrackerResponse struct {
	Interval int    `mapstructure:"interval"`
	Peers    []Peer `mapstructure:"peers"`
}

func FetchPeers(trackerUrl string, infoHash []byte, length int) ([]Peer, error) {
	req, err := http.NewRequest(http.MethodGet, trackerUrl, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	q.Add("info_hash", string(infoHash))
	q.Add("peer_id", ID)
	q.Add("port", "6881")
	q.Add("uploaded", "0")
	q.Add("downloaded", "0")
	q.Add("left", strconv.Itoa(length))

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

	os.WriteFile("testdata/tmp.torrent", body, 0o600)

	decoded, err := bencode.Unmarshal(string(body))
	if err != nil {
		return nil, err
	}

	var tr TrackerResponse
	if err = mapstructure.Decode(decoded, &tr); err != nil {
		return nil, err
	}

	return tr.Peers, nil
}
