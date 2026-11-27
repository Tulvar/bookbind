package chapters

import (
	"fmt"
	"time"
)

const maxSyntheticChapters = 10_000

func Synthetic(total, interval time.Duration) ([]Chapter, error) {
	if total <= 0 {
		return nil, fmt.Errorf("total duration is required for synthetic chapters")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("chapter interval must be positive")
	}

	count := (total-1)/interval + 1
	if count > maxSyntheticChapters {
		return nil, fmt.Errorf("chapter interval %s would create %d chapters (maximum %d)", interval, count, maxSyntheticChapters)
	}
	result := make([]Chapter, 0, int(count))
	for start := time.Duration(0); start < total; {
		end := start + min(interval, total-start)
		result = append(result, Chapter{
			Title: fmt.Sprintf("Chapter %03d", len(result)+1),
			Start: start,
			End:   end,
		})
		start = end
	}
	return result, nil
}
