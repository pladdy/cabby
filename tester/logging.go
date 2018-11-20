package tester

import (
	"bytes"
	"regexp"
	"strings"
)

// ErrorLog for service errors
type ErrorLog struct {
	Error string
	Level string
	Msg   string
	Time  string
}

// RequestLog from http service
type RequestLog struct {
	Time      string
	Level     string
	Msg       string
	ElapsedTs float64 `json:"elapsed_ts"`
	EndTs     int64   `json:"end_ts"`
	Method    string
	URL       string
}

// LastLog given a buffer of logs returns the last log statement
func LastLog(buf bytes.Buffer) string {
	logs := regexp.MustCompile("\n").Split(strings.TrimSpace(buf.String()), -1)
	return logs[len(logs)-1]
}
