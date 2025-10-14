package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	JOIN       string = "JOIN"
	LEAVE             = "LEAVE"
	MOVE              = "MOVE"
	TELEPORTED        = "TELEPORTED"
	COMPLETED         = "COMPLETED"
)

type StreamLog struct {
	TimeStamp time.Time
	Message   string
	LogType   string
}

type StreamLogs []StreamLog

// Implementing heap functionalities in Stream Logs
func (sl StreamLogs) Len() int {
	return len(sl)
}

func (sl StreamLogs) Less(i, j int) bool {
	return sl[i].TimeStamp.Before(sl[j].TimeStamp)
}

func (sl StreamLogs) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}

func (sl *StreamLogs) Push(lg StreamLog) {
	*sl = append(*sl, lg)
}

func (sl *StreamLogs) Pop() StreamLog {
	old := *sl
	n := len(old)

	item := old[n-1]
	*sl = old[:n-1]

	return item
}

// Implementing streamer
type Stream struct {
	Logs    StreamLogs // buffered heap
	MaxSize int
}

func NewStreamer() *Stream {
	maxStreams, err := strconv.Atoi(os.Getenv("MAX_STREAMS"))
	if err != nil {
		log.Fatalf("error while parsing MAX_STREAMS | error: %v\n", err)
	}
	return &Stream{
		Logs:    StreamLogs{},
		MaxSize: maxStreams,
	}
}

func (stream *Stream) Push(lg StreamLog) {
	// Checking if the stream size has reached its limit
	if len(stream.Logs) == stream.MaxSize {
		stream.Logs.Pop()
	}

	stream.Logs.Push(lg)
}
