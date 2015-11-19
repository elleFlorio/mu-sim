package network

import "time"

type Request struct {
	ID         string
	From       string
	To         string
	Counter    int
	Start      time.Time
	ExecTimeMs float64
}
