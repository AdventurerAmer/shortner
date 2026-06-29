package snowflake

import (
	"strings"
	"sync"
	"time"
)

const (
	Epoch = 1704067200000 // 2024-01-01 00:00:00 UTC

	TimestampBits   = 41
	SequenceNumBits = 23 // 8,388,608 IDs per ms
	MaxSequenceNum  = (1 << SequenceNumBits) - 1
)

type Generator struct {
	mu            sync.Mutex
	shard         string
	lastTimestamp int64
	sequenceNum   int64
}

func New(shard string) *Generator {
	return &Generator{
		shard:         shard,
		lastTimestamp: time.Now().UnixMilli(),
	}
}

func (g *Generator) Next() string {
	id := g.NextInt64()
	return g.shard + toBase62(id)
}

func (g *Generator) NextInt64() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := time.Now().UnixMilli()
	if timestamp < g.lastTimestamp {
		timestamp = g.lastTimestamp
	}
	if timestamp == g.lastTimestamp {
		g.sequenceNum = (g.sequenceNum + 1) & MaxSequenceNum
		if g.sequenceNum == 0 {
			// Wait until next millisecond
			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequenceNum = 0
	}

	g.lastTimestamp = timestamp
	id := (timestamp - Epoch) | g.sequenceNum
	return id
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
