package utils

import (
	"crypto/rand"
	"math/big"
)

var charset = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateRandomJoinCode() string {
	b := make([]rune, 9)
	for i := 0; i < 9; i++ {
		if i == 4 {
			b[i] = '-'
			continue
		}

		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic("failed to generate random number: " + err.Error())
		}

		b[i] = charset[num.Int64()]
	}

	return string(b)
}