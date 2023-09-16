package config

import "time"

type Config struct {
	Path       string
	Baud       int
	Size       int
	Parity     string
	Timeout    time.Duration
	StopBit    int
	Write      string
	Once       bool
	WriteDelay time.Duration
	TestRead   string
}
