package crypto

import (
	"fmt"
	"testing"
)

func TestGetRandomNumber(t *testing.T) {

	// Lets get a 6 digit number
	random := GetRandomInt(100000, 999999)

	if random < 100000 || random > 999999 {
		t.Fatal("Number is outside of desired range")
	}

	fmt.Printf("Random is: %v\n", random)
}

func TestRandomSecret(t *testing.T) {

	secret := RandomSecret(0)

	if secret == "" {
		t.Fatal("Secret is empty")
	}

	if len(secret) != 16 {
		t.Fatal("Secret is too short by default")
	}

	fmt.Printf("Random is: %v\n", secret)
}

func Test32CharRandomSecret(t *testing.T) {

	secret := RandomSecret(40)

	fmt.Printf("Random is: %v Len: %v\n", secret, len(secret))

	if secret == "" {
		t.Fatal("Secret is empty")
	}

	if len(secret) != 64 {
		t.Fatal("Secret is not 64")
	}

	fmt.Printf("Random is: %v\n", secret)
}
