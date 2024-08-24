package message

type Bitfield []byte

// HasPiece check for existence of a piece given a bitfield payload
func (bf Bitfield) HasPiece(pieceIndex int) bool {
	byteIdx := pieceIndex / 8
	if byteIdx < 0 || byteIdx >= len(bf) {
		return false
	}
	// "7 - " since the bitfield is in big endian
	bitIdx := 7 - (pieceIndex % 8)

	return bf[byteIdx]&(1<<bitIdx) != 0
}
