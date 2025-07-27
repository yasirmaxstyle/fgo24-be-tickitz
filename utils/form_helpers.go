package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GetStringField(form map[string][]string, key string) *string {
	if val, ok := form[key]; ok && len(val) > 0 {
		s := val[0]
		return &s
	}
	return nil
}

func GetIntField(form map[string][]string, key string) (*int, error) {
	if val, ok := form[key]; ok && len(val) > 0 {
		i, err := strconv.Atoi(val[0])
		if err != nil {
			return nil, fmt.Errorf("invalid int value for %s", key)
		}
		return &i, nil
	}
	return nil, nil
}

func GetDateField(form map[string][]string, key string) (*time.Time, error) {
	if val, ok := form[key]; ok && len(val) > 0 {
		t, err := time.Parse("2006-01-02", val[0])
		if err != nil {
			return nil, fmt.Errorf("invalid date format for %s", key)
		}
		return &t, nil
	}
	return nil, nil
}

func GetIntArray(form map[string][]string, key string) (*[]int, error) {
	if val, ok := form[key]; ok && len(val) > 0 {
		parts := strings.Split(val[0], ",")
		var result []int
		for _, s := range parts {
			i, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return nil, fmt.Errorf("invalid value in array: %s", s)
			}
			result = append(result, i)
		}
		return &result, nil
	}
	return &[]int{}, nil
}
