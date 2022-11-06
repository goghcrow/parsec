package example

import (
	"errors"
	"strconv"
	"strings"
)

func parseInt(s string) (int64, error) {
	if strings.HasPrefix(s, "0x") {
		n, err := strconv.ParseInt(s[2:], 16, 64)
		if err == nil {
			return n, nil
		}
	}
	if strings.HasPrefix(s, "0b") {
		n, err := strconv.ParseInt(s[2:], 2, 64)
		if err == nil {
			return n, nil
		}
	}
	if strings.HasPrefix(s, "0o") {
		n, err := strconv.ParseInt(s[2:], 8, 64)
		if err == nil {
			return n, nil
		}
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return n, nil
	}

	return 0, errors.New("invalid Int: " + s)
}
