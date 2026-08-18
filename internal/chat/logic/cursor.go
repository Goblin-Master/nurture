package logic

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func parseCursor(cursor string) (int64, string, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return 0, "", ErrInvalidCursor
	}
	ctime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || ctime <= 0 {
		return 0, "", ErrInvalidCursor
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return 0, "", ErrInvalidCursor
	}
	return ctime, parts[1], nil
}

func buildCursor(ctime int64, messageID string) string {
	return strconv.FormatInt(ctime, 10) + "|" + messageID
}

func parseDiscoverCursor(cursor string) (string, string, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return "", "", ErrInvalidCursor
	}
	sortKey := parts[0]
	if len(sortKey) != 32 {
		return "", "", ErrInvalidCursor
	}
	for _, c := range sortKey {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return "", "", ErrInvalidCursor
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return "", "", ErrInvalidCursor
	}
	return sortKey, parts[1], nil
}
