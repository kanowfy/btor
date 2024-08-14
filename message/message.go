package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageID byte

const (
	MessageChoke MessageID = iota
	MessageUnchoke
	MessageInterested
	MessageUninterested
	MessageHave
	MessageBitfield
	MessageRequest
	MessagePiece
	MessageCancel
)

type Message struct {
	ID      MessageID
	Payload []byte
}

func New(id MessageID, payload []byte) *Message {
	return &Message{
		ID:      id,
		Payload: payload,
	}
}

func (m *Message) Serialize() []byte {
	length := 5 + len(m.Payload)
	buf := make([]byte, length)
	binary.BigEndian.PutUint32(buf[0:4], uint32(length))
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)

	return buf
}

// Read reads a Message from an io.Reader, returns nil error for keep-alive message
func Read(r io.Reader) (*Message, error) {
	// reads in the length
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	// 0 len message means keep-alive message
	if length == 0 {
		return nil, nil
	}

	msgBuf := make([]byte, length)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		return nil, err
	}

	message := new(Message)

	switch msgBuf[0] {
	case 0:
		message.ID = MessageChoke
	case 1:
		message.ID = MessageUnchoke
	case 2:
		message.ID = MessageInterested
	case 3:
		message.ID = MessageUninterested
	case 4:
		message.ID = MessageHave
	case 5:
		message.ID = MessageBitfield
	case 6:
		message.ID = MessageRequest
	case 7:
		message.ID = MessagePiece
	case 8:
		message.ID = MessageCancel
	default:
		return nil, fmt.Errorf("invalid message.ID: %d", msgBuf[0])
	}

	message.Payload = msgBuf[1:]

	return message, nil
}

func NewRequest(pieceIndex, blockOffset, blockLength int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(payload[4:8], uint32(blockOffset))
	binary.BigEndian.PutUint32(payload[8:12], uint32(blockLength))
	return &Message{
		ID:      MessageRequest,
		Payload: payload,
	}
}

func ParsePiece(msg *Message, buf []byte, pieceIndex int) (int, error) {
	index := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if index != pieceIndex {
		return 0, fmt.Errorf("unexpected piece index, want %d, got %d", pieceIndex, index)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	data := msg.Payload[8:]
	copy(buf[begin:], data)
	return len(data), nil
}
