package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseChapterInterval(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	if duration, err := time.ParseDuration(value); err == nil {
		if duration <= 0 {
			return 0, fmt.Errorf("chapter interval must be positive")
		}
		return duration, nil
	}

	minutes, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid chapter interval %q", value)
	}
	if minutes <= 0 {
		return 0, fmt.Errorf("chapter interval must be positive")
	}
	return time.Duration(minutes) * time.Minute, nil
}
