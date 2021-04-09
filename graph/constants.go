package graph

import "time"

const (
	clientTimeout = 10 * time.Second // client req. timeout
	maxSummaryLen = 40               // Max. length for an auto-generated description
)
