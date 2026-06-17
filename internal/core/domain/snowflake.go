package domain

import (
	"strings"
	"sync"
	"time"
)

const (
	sequenceBits = 23 // 8,388,608 IDs per ms
	maxSequence  = (1 << sequenceBits) - 1
	customEpoch  = 1704067200000 // 2024-01-01 00:00:00 UTC
)

type Snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequence      int64
}

func NewSnowflake() *Snowflake {
	return &Snowflake{
		lastTimestamp: time.Now().UnixMilli(),
	}
}

func (g *Snowflake) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := time.Now().UnixMilli()
	if timestamp < g.lastTimestamp {
		timestamp = g.lastTimestamp
	}
	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			// Wait until next millisecond
			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = timestamp
	id := (timestamp - customEpoch) | g.sequence
	return id
}

func (g *Snowflake) NextBase62(prefix string) string {
	return prefix + toBase62(g.Next())
}

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func toBase62(num int64) string {
	if num == 0 {
		return "0"
	}

	var encoded strings.Builder
	for num > 0 {
		encoded.WriteByte(base62Alphabet[num%62])
		num /= 62
	}

	// Reverse the string
	runes := []rune(encoded.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
