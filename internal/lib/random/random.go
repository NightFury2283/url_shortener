package random

import (
	"math/rand"
)

func GenerateRandomString(length int) string {
	stroke := ""

	words := "abcdefghijklmnopqrstuvwxyz"

	for i := 0; i < length; i++ {
		stroke += string(words[rand.Intn(len(words))])
	}
	return stroke
}

//TODO: add check in bd for existing strings and regenerate if exists