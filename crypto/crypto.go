package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func CBCDecrypt(encryptedBytes []byte) ([]byte, error) {
	key := []byte("a very very very very secret key") // 32 bytes

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(encryptedBytes) < aes.BlockSize {
		return nil, errors.New("encryptedBytes too short")
	}
	iv := encryptedBytes[:aes.BlockSize]
	encryptedBytes = encryptedBytes[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(encryptedBytes)%aes.BlockSize != 0 {
		return nil, errors.New("encryptedBytes is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(encryptedBytes, encryptedBytes)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point. For an example, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
	// critical to note that ciphertexts must be authenticated (i.e. by
	// using crypto/hmac) before being decrypted in order to avoid creating
	// a padding oracle.

	return encryptedBytes, nil
}

// Encrypt a slice of bytes
func CBCEncrypt(bytesToEncrypt []byte) ([]byte, error) {
	key := []byte("a very very very very secret key") // 32 bytes

	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	if len(bytesToEncrypt)%aes.BlockSize != 0 {
		return nil, errors.New("bytesToEncrypt is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	encryptedBytes := make([]byte, aes.BlockSize+len(bytesToEncrypt))
	iv := bytesToEncrypt[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(bytesToEncrypt[aes.BlockSize:], encryptedBytes)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.

	return encryptedBytes, nil
}

func AesCfbDecrypt(encryptedBytes []byte) ([]byte, error) {
	key := []byte("a very very very very secret key") // 32 bytes

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(encryptedBytes) < aes.BlockSize {
		return nil, errors.New("Data to encrypt is too small")
	}
	iv := encryptedBytes[:aes.BlockSize]
	encryptedBytes = encryptedBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(encryptedBytes, encryptedBytes)
	// fmt.Printf("%s", ciphertext)
	// Output: some plaintext

	return encryptedBytes, nil
}

func AesCfbEncrypt(bytesToEncrypt []byte) ([]byte, error) {
	key := []byte("a very very very very secret key") // 32 bytes

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(bytesToEncrypt))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], bytesToEncrypt)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	return ciphertext, nil
}
