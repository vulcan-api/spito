package vrctFs

import (
	"fmt"
	"math/rand"
)

func randomLetters(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}

func pathMustBeAbsolute(path string) (string, error) {
	if path[0] != '/' {
		return "", fmt.Errorf("Path must be absolute")
	}

	return path, nil
}
