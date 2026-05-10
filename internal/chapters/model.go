package chapters

import "time"

type Chapter struct {
	Title string
	Start time.Duration
	End   time.Duration
}
