package tester

import (
	"log"
	"os"
)

var (
	// Info logs
	Info = log.New(os.Stderr, "Test INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	// Warn logs
	Warn = log.New(os.Stderr, "Test WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	// Error logs
	Error = log.New(os.Stderr, "Test ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)
