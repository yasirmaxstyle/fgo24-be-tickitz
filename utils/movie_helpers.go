package utils

import (
	"strings"
)

func SplitFullName(fullname string) (firstName, lastName string) {
	parts := strings.Fields(fullname)
	if len(parts) == 0 {
		return "", ""
	} else if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}
