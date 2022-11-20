package common

import "strings"

func RightSplit(s string, sep string) []string {
	var split []string

	if !strings.Contains(s, sep) {
		return []string{s}
	}

	fields := strings.Split(s, sep)
	left := fields[:len(fields)-1]
	leftJoined := strings.Join(left, sep)
	split = append(split, leftJoined)

	if len(fields) > 1 {
		split = append(split, fields[len(fields)-1])
	}

	return split
}
