package crypto

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	mathrand "math/rand"
)

// Get a reandom number
func GetRandomInt(min, max int) int {

	// Generate a Crypto random seed from the OS
	// We should not use the time as the seed as this will lead to predicatable PINs
	var n int64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	mathrand.Seed(n)

	// Now get a number from the range desired
	return mathrand.Intn(max-min) + min
}

// Generate a Random secret encoded as a b32 string
// If the length is <= 0, a default length of 10 bytes will
// be used, which will generate a secret of length 16.
func RandomSecret(length int) string {
	if length <= 0 {
		length = 10
	}

	// Get a random based on a random int.  Based off OS not based on Time.
	rnd := mathrand.New(mathrand.NewSource(int64(GetRandomInt(100000, 999999))))

	secret := make([]byte, length)
	for i, _ := range secret {
		secret[i] = byte(rnd.Int31() % 256)
	}
	return base32.StdEncoding.EncodeToString(secret)
}
