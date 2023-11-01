package jitter

import (
	"math/rand"
	"time"
)

func Seed() {
	rand.Seed(time.Now().UnixNano())
}

func GetRandomDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Int63n(max.Nanoseconds()))
}

func Stringify(jitter time.Duration) string {
	trunc := time.Duration(10000000)
	return jitter.Truncate(trunc).String()
}