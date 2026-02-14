package nanoid

import (
	"crypto/rand"
	"math/big"
)

const (
	// Alphabet for URL-safe short codes
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// DefaultSize is the default length of generated IDs
	DefaultSize = 8
)

func Generate(size int) (string, error) {
	if size <= 0 {
		size = DefaultSize
	}

	bytes := make([]byte, size)
	alphabetLen := big.NewInt(int64(len(alphabet)))

	for i := 0; i < size; i++ {
		num, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		bytes[i] = alphabet[num.Int64()]
	}

	return string(bytes), nil
}

func MustGenerate(size int) string {
	id, err := Generate(size)
	if err != nil {
		panic(err)
	}
	return id
}

func New() (string, error) {
	return Generate(DefaultSize)
}
