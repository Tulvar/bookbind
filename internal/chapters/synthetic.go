package chapters

import (
	"fmt"
	"time"
)

func Synthetic(total, interval time.Duration) ([]Chapter, error) {
	if total <= 0 {
		return nil, fmt.Errorf("total duration is required for synthetic chapters")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("chapter interval must be positive")
	}

	result := make([]Chapter, 0, int(total/interval)+1)
	for start := time.Duration(0); start < total; start += interval {
		end := start + interval
		if end > total {
			end = total
		}
		result = append(result, Chapter{
			Title: fmt.Sprintf("Chapter %03d", len(result)+1),
			Start: start,
			End:   end,
		})
	}
	return result, nil
}
