package crypto

import (
	"bytes"
	"crypto"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
	"io"
	"log"
	"os"
)

// Default encryption settings (No compression done by pgp)
var pgpConfig = &packet.Config{
	DefaultHash:            crypto.SHA256,
	DefaultCipher:          packet.CipherAES256,
	DefaultCompressionAlgo: packet.CompressionNone,
}

// AwsCryptoProvider is an implementation of encryption using AWS KMS
type OpenPGPCryptoProvider struct {
	// publicEntityList - Public Key
	publicEntityList openpgp.EntityList
	// privateEntityList - Private Key
	privateEntityList openpgp.EntityList
}

// NewAwsCryptoProvider
func NewOpenPGPCryptoProvider() (OpenPGPCryptoProvider, error) {

	log.Println("Using OpenPGPCryptoProvider for encryption...")

	publicKeyPath := os.Getenv("BLOCKER_PGP_PUBLICKEY")
	if publicKeyPath == "" {
		panic("You must specify a public pgp key.  Set env: BLOCKER_PGP_PUBLICKEY")
	}

	privateKeyPath := os.Getenv("BLOCKER_PGP_PRIVATEKEY")
	if publicKeyPath == "" {
		panic("You must specify a private pgp key.  Set env: BLOCKER_PGP_PRIVATEKEY")
	}

	log.Printf("Reading public key from %s\n", publicKeyPath)
	publicKey, err := os.Open(publicKeyPath)
	defer publicKey.Close()
	if err != nil {
		panic(err)
	}

	publicEntityList, err := openpgp.ReadArmoredKeyRing(publicKey)
	if err != nil {
		panic(err)
	}

	log.Printf("Reading private key from %s\n", privateKeyPath)
	privateKey, err := os.Open(privateKeyPath)
	defer privateKey.Close()
	if err != nil {
		panic(err)
	}
	privateEntityList, err := openpgp.ReadArmoredKeyRing(privateKey)
	if err != nil {
		panic(err)
	}

	return OpenPGPCryptoProvider{publicEntityList: publicEntityList, privateEntityList: privateEntityList}, nil
}

// Decrypt decrypts data that has been encrypted and compressed
func (p OpenPGPCryptoProvider) Decrypt(data []byte) ([]byte, error) {
	dataBuffer := bytes.NewReader(data)
	md, err := openpgp.ReadMessage(dataBuffer, p.privateEntityList, nil, pgpConfig)
	if err != nil {
		return nil, err
	}

	// Read all the converted data...
	var b bytes.Buffer
	b.ReadFrom(md.UnverifiedBody)
	return b.Bytes(), nil
}

// Encrypt - Encrypts the data
func (p OpenPGPCryptoProvider) Encrypt(data []byte) ([]byte, error) {
	encryptedBuffer := &bytes.Buffer{}
	dataBuffer := bytes.NewReader(data)

	// Call openpgp encrypt with default settings
	pgpWriter, err := openpgp.Encrypt(encryptedBuffer, p.publicEntityList, nil, nil, pgpConfig)
	if err != nil {
		return nil, err
	}

	// Encrypt streams
	io.Copy(pgpWriter, dataBuffer)

	// Close the encryption stream
	if err := pgpWriter.Close(); err != nil {
		return nil, err
	}

	// return the encrypted bytes
	return encryptedBuffer.Bytes(), nil
}
