package util

import (
	"math/rand"
	"strings"
	"time"
)

var Rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var alphabet = "abcdefghijklmnopqrstuvwxyz"

// generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// generates a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// generate a random amount of money between 0 and 1000
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// generate a random currency
func RandomCurrency() string {
	currencies := []string{USD, EUR, CAD}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

// RandomAccountId generates a random account id from 1 to 27
func RandomAccountId() int64 {
	return RandomInt(1, 27)
}

func RandomAmount() int64 {
	return RandomInt(0, 100)
}
