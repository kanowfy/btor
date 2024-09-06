package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/jackpal/bencode-go"
	"github.com/mitchellh/mapstructure"
)

type Peer struct {
	ID   string `mapstructure:"peer id"`
	IP   string `mapstructure:"ip"`
	Port uint16 `mapstructure:"port"`
}

// Fetch sends a GET request to a tracker endpoint, parses the response and returns the peers
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
	q.Add("compact", "1")

	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoded, err := bencode.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	var peers []Peer

	// some tracker will not return compact peer representation, this will support both compact and non-compact ones
	peerResp := decoded.(map[string]interface{})["peers"]
	switch v := peerResp.(type) {
	case string:
		peers, err = parseCompactPeers([]byte(v))
		if err != nil {
			return nil, fmt.Errorf("failed to parse peers: %v", err)
		}
	case []interface{}:
		mapstructure.Decode(v, &peers)
	default:
		return nil, fmt.Errorf("invalid peer representation")
	}

	return peers, nil
}

func parseCompactPeers(peerString []byte) ([]Peer, error) {
	numPeers := len(peerString) / 6
	peers := make([]Peer, numPeers)
	for i := range numPeers {
		peers[i] = Peer{
			IP:   net.IP(peerString[i*6 : i*6+4]).String(),
			Port: binary.BigEndian.Uint16(peerString[i*6+4 : i*6+6]),
		}
	}
	return peers, nil
}
