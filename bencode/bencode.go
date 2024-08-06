package bencode

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"unicode"
)

func Decode(bencode string, pos int) (interface{}, int, error) {
	switch {
	case unicode.IsDigit(rune(bencode[pos])):
		return DecodeString(bencode, pos)
	case bencode[pos] == 'i':
		return DecodeInteger(bencode, pos)
	case bencode[pos] == 'l':
		return DecodeList(bencode, pos)
	case bencode[pos] == 'd':
		return DecodeDict(bencode, pos)
	default:
		return nil, 0, fmt.Errorf("invalid character at pos %d", pos)
	}
}

func DecodeString(bencode string, pos int) (string, int, error) {
	var colonIndex int

	for i := pos; i < len(bencode); i++ {
		if bencode[i] == ':' {
			colonIndex = i
			break
		}
	}

	if colonIndex == 0 {
		return "", 0, fmt.Errorf("invalid bencoded value: %q", bencode)
	}

	lengthStr := bencode[pos:colonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	result := bencode[colonIndex+1 : colonIndex+1+length]

	return result, colonIndex + 1 + length, nil
}

func DecodeInteger(bencode string, pos int) (int, int, error) {
	pos += 1
	var end int

	for i := pos; i < len(bencode); i++ {
		if bencode[i] == 'e' {
			end = i
			break
		}
	}

	if end == 0 {
		return 0, 0, fmt.Errorf("invalid bencoded value: %q", bencode)
	}

	valueStr := bencode[pos:end]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, 0, err
	}

	return value, end + 1, nil
}

func DecodeList(bencode string, pos int) ([]interface{}, int, error) {
	pos += 1
	list := make([]interface{}, 0)

	for pos < len(bencode) && bencode[pos] != 'e' {
		decoded, newPos, err := Decode(bencode, pos)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, decoded)
		pos = newPos
	}

	return list, pos + 1, nil
}

func DecodeDict(bencode string, pos int) (map[string]interface{}, int, error) {
	pos += 1
	dict := make(map[string]interface{})

	for pos < len(bencode) && bencode[pos] != 'e' {
		key, posKey, err := DecodeString(bencode, pos)
		if err != nil {
			return nil, 0, err
		}

		pos = posKey
		val, posVal, err := Decode(bencode, pos)
		if err != nil {
			return nil, 0, err
		}
		dict[key] = val
		pos = posVal
	}

	return dict, pos + 1, nil
}

// TODO: refactor encoding
func EncodeDict(obj map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('d')

	// sort keys for deterministic map access
	i := 0
	keys := make([]string, len(obj))
	for k := range obj {
		keys[i] = k
		i++
	}
	slices.Sort(keys)

	for _, k := range keys {
		l := strconv.Itoa(len(k))
		buf.WriteString(l)
		buf.WriteByte(':')
		buf.WriteString(k)

		switch val := obj[k].(type) {
		case string:
			l := strconv.Itoa(len(val))
			buf.WriteString(l)
			buf.WriteByte(':')
			buf.WriteString(val)
		case int:
			buf.WriteByte('i')
			i := strconv.Itoa(val)
			buf.WriteString(i)
			buf.WriteByte('e')
		case map[string]interface{}:
			encoded, err := EncodeDict(val)
			if err != nil {
				return nil, err
			}
			buf.Write(encoded)
		default:
			return nil, fmt.Errorf("unsupported value: %v", val)
		}
	}

	buf.WriteByte('e')
	return buf.Bytes(), nil
}
