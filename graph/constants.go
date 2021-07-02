package graph

import "time"

const (
	dialTimeout           = 30 * time.Second
	KeepAliveTimeout      = 30 * time.Second
	clientTimeout         = 10 * time.Second // client req. timeout
	maxIdleConnsPerHost   = 20
	idleConnTimeout       = 90 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	expectContinueTimeout = 1 * time.Second
)
