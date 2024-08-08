package bencode

import (
	"fmt"
	"slices"
	"strconv"
	"unicode"
)

func Unmarshal(bencode string) (interface{}, error) {
	res, _, err := Decode(bencode, 0)
	return res, err
}

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

func Marshal(value any) ([]byte, error) {
	return Encode(value, []byte{})
}

// TODO: see if can use bytes.Buffer to read and write to the same object
func Encode(value any, parts []byte) ([]byte, error) {
	switch v := value.(type) {
	case int:
		return EncodeInteger(v, parts), nil
	case string:
		return EncodeString(v, parts), nil
	case []interface{}:
		return EncodeList(v, parts), nil
	case map[string]interface{}:
		return EncodeDict(v, parts), nil
	default:
		return nil, fmt.Errorf("invalid value: %v", v)
	}
}

func EncodeInteger(num int, parts []byte) []byte {
	parts = append(parts, 'i')
	parts = append(parts, strconv.Itoa(num)...)
	parts = append(parts, 'e')
	return parts
}

func EncodeString(str string, parts []byte) []byte {
	parts = append(parts, strconv.Itoa(len(str))...)
	parts = append(parts, ':')
	parts = append(parts, []byte(str)...)
	return parts
}

func EncodeList(list []interface{}, parts []byte) []byte {
	parts = append(parts, 'l')
	for _, elem := range list {
		res, err := Encode(elem, parts)
		if err != nil {
			return nil
		}

		parts = res
	}

	parts = append(parts, 'e')
	return parts
}

func EncodeDict(dict map[string]interface{}, parts []byte) []byte {
	parts = append(parts, 'd')
	i := 0
	keys := make([]string, len(dict))
	for k := range dict {
		keys[i] = k
		i++
	}
	slices.Sort(keys)

	for _, k := range keys {
		parts = EncodeString(k, parts)

		valPart, err := Encode(dict[k], parts)
		if err != nil {
			return nil
		}

		parts = valPart
	}

	parts = append(parts, 'e')
	return parts
}
