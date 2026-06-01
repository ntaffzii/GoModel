package streaming

import (
	"io"
	"strconv"
	"unicode/utf8"
)

type JSONUnwrapperReader struct {
	io.ReadCloser
	checked     bool
	isJSON      bool
	state       int // 0: string, 1: escape, 2: unicode
	unicodeBuf  [4]byte
	unicodeLen  int
	eofReached  bool
	decoded     []byte
}

func NewJSONUnwrapperReader(rc io.ReadCloser) io.ReadCloser {
	return &JSONUnwrapperReader{ReadCloser: rc}
}

const (
	stateString = iota
	stateEscape
	stateUnicode
)

func (u *JSONUnwrapperReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if !u.checked {
		var firstChunk [4096]byte
		nRead, readErr := u.ReadCloser.Read(firstChunk[:])
		if nRead == 0 {
			if readErr != nil {
				return 0, readErr
			}
			return 0, nil
		}
		u.checked = true
		if firstChunk[0] == '"' {
			u.isJSON = true
			u.state = stateString
			u.decodeBytes(firstChunk[1:nRead])
		} else {
			u.isJSON = false
			u.decoded = append(u.decoded, firstChunk[:nRead]...)
		}
		if readErr != nil && readErr != io.EOF {
			return len(u.decoded), readErr
		}
	}

	if !u.isJSON {
		if len(u.decoded) > 0 {
			nCopied := copy(p, u.decoded)
			u.decoded = u.decoded[nCopied:]
			return nCopied, nil
		}
		return u.ReadCloser.Read(p)
	}

	if len(u.decoded) > 0 {
		nCopied := copy(p, u.decoded)
		u.decoded = u.decoded[nCopied:]
		return nCopied, nil
	}

	if u.eofReached {
		return 0, io.EOF
	}

	var raw [4096]byte
	nRaw, rawErr := u.ReadCloser.Read(raw[:])
	if nRaw == 0 {
		if rawErr != nil {
			if rawErr == io.EOF {
				u.eofReached = true
			}
			return 0, rawErr
		}
		return 0, nil
	}

	u.decodeBytes(raw[:nRaw])

	if len(u.decoded) > 0 {
		nCopied := copy(p, u.decoded)
		u.decoded = u.decoded[nCopied:]
		return nCopied, nil
	}

	if rawErr == io.EOF {
		u.eofReached = true
		return 0, io.EOF
	}

	return 0, nil
}

func (u *JSONUnwrapperReader) decodeBytes(raw []byte) {
	rawIdx := 0
	nRaw := len(raw)
	for rawIdx < nRaw {
		b := raw[rawIdx]
		rawIdx++

		switch u.state {
		case stateString:
			switch b {
			case '\\':
				u.state = stateEscape
			case '"':
				u.eofReached = true
				rawIdx = nRaw
			default:
				u.decoded = append(u.decoded, b)
			}

		case stateEscape:
			switch b {
			case 'n':
				u.decoded = append(u.decoded, '\n')
				u.state = stateString
			case 'r':
				u.decoded = append(u.decoded, '\r')
				u.state = stateString
			case 't':
				u.decoded = append(u.decoded, '\t')
				u.state = stateString
			case 'b':
				u.decoded = append(u.decoded, '\b')
				u.state = stateString
			case 'f':
				u.decoded = append(u.decoded, '\f')
				u.state = stateString
			case '"':
				u.decoded = append(u.decoded, '"')
				u.state = stateString
			case '\\':
				u.decoded = append(u.decoded, '\\')
				u.state = stateString
			case '/':
				u.decoded = append(u.decoded, '/')
				u.state = stateString
			case 'u':
				u.state = stateUnicode
				u.unicodeLen = 0
			default:
				u.decoded = append(u.decoded, '\\', b)
				u.state = stateString
			}

		case stateUnicode:
			u.unicodeBuf[u.unicodeLen] = b
			u.unicodeLen++
			if u.unicodeLen == 4 {
				val, parseErr := strconv.ParseUint(string(u.unicodeBuf[:]), 16, 16)
				u.state = stateString
				if parseErr == nil {
					r := rune(val)
					var buf [utf8.UTFMax]byte
					runeLen := utf8.EncodeRune(buf[:], r)
					u.decoded = append(u.decoded, buf[:runeLen]...)
				} else {
					u.decoded = append(u.decoded, '\\', 'u')
					u.decoded = append(u.decoded, u.unicodeBuf[:]...)
				}
			}
		}
	}
}
