package utils

import (
	"math/rand"
	"time"
)

func GenerateRandomNumber() int64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Int63()
}
