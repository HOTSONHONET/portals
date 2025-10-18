package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	JOIN       string = "JOIN"
	LEAVE      string = "LEAVE"
	MOVE       string = "MOVE"
	TELEPORTED string = "TELEPORTED"
	COMPLETED  string = "COMPLETED"
)

type StreamLog struct {
	TimeStamp time.Time
	Message   string
	LogType   string
}

/* Ring Buffer: Max Heap by Timestamp */
type RingBuffer struct {
	data       []StreamLog
	head, tail int
	full       bool
	size       int
}

func (r *RingBuffer) Add(log StreamLog) {
	r.data[r.tail] = log
	r.tail = (r.tail + 1) % r.size
	if r.full {
		r.head = (r.head + 1) % r.size
	} else if r.head == r.tail {
		r.full = true
	}
}

func (r *RingBuffer) GetLatestLogs() []StreamLog {
	var res []StreamLog
	// if it is empty
	if !r.full && r.head == r.tail {
		return res
	}

	i := r.tail - 1
	if i < 0 {
		i = r.size - 1
	}

	for {
		res = append(res, r.data[i])
		if i == r.head && r.full {
			break
		}

		if !r.full && i == 0 {
			break
		}

		i--

		if i < 0 {
			i = r.size - 1
		}

		if !r.full && i == r.tail-1 {
			break
		}
	}

	return res
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]StreamLog, size),
		size: size,
	}
}

// Implementing streamer
type Stream struct {
	Logs *RingBuffer
	mu   sync.Mutex
}

func NewStreamer() *Stream {
	maxStreams, err := strconv.Atoi(os.Getenv("MAX_STREAMS"))
	if err != nil {
		log.Fatalf("error while parsing MAX_STREAMS | error: %v\n", err)
	}
	s := &Stream{
		Logs: NewRingBuffer(maxStreams),
	}
	return s
}

func (s *Stream) Push(log StreamLog) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Logs.Add(log)
}

func (s *Stream) GetLogs() []StreamLog {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Logs.GetLatestLogs()
}
