package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	// "crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	// "math/big"
	"os"
	"path/filepath"
	// "time"
)

// Path to the certificate
// var certifcatePath = filepath.Join(os.TempDir(), "blocks", "cert.pem")

// Path to the private key
var keyPath = filepath.Join(os.TempDir(), "blocks", "key.pem")

// Path to the encrypted aes key
var aesKeyPath = filepath.Join(os.TempDir(), "blocks", "aes.key")

// The key to be used to encrypt and decrypt when using RSA encryption
var rsaEncryptionChipher RsaChipher

// Structure for encryption chipher
type RsaChipher struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// The AES key used for AES encryption
var aesEncryptionKey AesKey

// Structure to hold unencrypted AES key
type AesKey struct {
	key []byte
}

func init() {
	LoadOrGenerateRsaKey()
	GetAesSecret()
}

// Load or Generate a RSA certiciate
func LoadOrGenerateRsaKey() {

	// Read key
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err == nil {
		// Get private key
		block, _ := pem.Decode(keyBytes)
		privatekey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

		// Set object
		rsaEncryptionChipher = RsaChipher{privatekey, &privatekey.PublicKey}

		// We are done
		return
	}

	// No load of existing key.  Generate a new one.
	GenerateRsaKey()
}

// Generate a new key
func GenerateRsaKey() {

	// Generate a 256 bit private key for use with the encryption
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
		return
	}

	/*now := time.Now()

	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   "Acme Encryption Certificate",
			Organization: []string{"Acme Co"},
		},
		NotBefore: now.Add(-5 * time.Minute).UTC(),
		NotAfter:  now.AddDate(1, 0, 0).UTC(), // valid for 1 year.

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
		return
	}

	certOut, err := os.Create(certifcatePath)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
		return
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("written cert.pem\n")*/

	depositoryDir := filepath.Join(os.TempDir(), "blocks")

	err = os.Mkdir(depositoryDir, 0777)
	if err != nil && !os.IsExist(err) {
		panic("Unable to create directory: " + err.Error())
	}

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Print("failed to open key.pem for writing:", err)
		return
	}

	marashelledPrivateKeyBytes := x509.MarshalPKCS1PrivateKey(priv)

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: marashelledPrivateKeyBytes})
	keyOut.Close()

	log.Print("Wrote Certificate to disk.")

	// Now set object
	rsaEncryptionChipher = RsaChipher{priv, &priv.PublicKey}
}

// Encrypt data using RSA and a public key
func RsaEncrypt(bytesToEncrypt []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, rsaEncryptionChipher.PublicKey, bytesToEncrypt)
}

// Decrypt data using RSA and a private key
func RsaDecrypt(encryptedBytes []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, rsaEncryptionChipher.PrivateKey, encryptedBytes)
}

// Create a new Aes Secret
func GenerateAesSecret() []byte {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)
	return key
}

// Get the AES secret to be used for encryption
func GetAesSecret() (err error) {
	// Read key
	keyBytes, err := ioutil.ReadFile(aesKeyPath)
	if err == nil {
		key, _ := RsaDecrypt(keyBytes)
		aesEncryptionKey = AesKey{key}
	}

	// Create new Aes Secret
	newAesKey := GenerateAesSecret()

	// Encrypt the key for later use
	encryptedKey, err := RsaEncrypt(newAesKey)
	if err != nil {
		log.Println(fmt.Sprintf("Error writing file : %v", err))
		return err
	}

	// Save encrypted key to disk
	err = ioutil.WriteFile(aesKeyPath, encryptedKey, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("Error writing file : %v", err))
		return err
	}

	aesEncryptionKey = AesKey{newAesKey}

	return nil
}

// Hex to bytes
func hex2Bytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// Bytes to hex
func encodeHex(bytes []byte) string {
	return fmt.Sprintf("%x", bytes)
}

// Encrpyt data using AES with the CFB chipher mode
func AesCfbDecrypt(encryptedBytes []byte) ([]byte, error) {
	// key := []byte("a very very very very secret key") // 32 bytes

	block, err := aes.NewCipher(aesEncryptionKey.key)
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

// Encrpyt data using AES with the CFB chipher mode
func AesCfbEncrypt(bytesToEncrypt []byte) ([]byte, error) {
	// key := []byte("a very very very very secret key") // 32 bytes

	block, err := aes.NewCipher(aesEncryptionKey.key)
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
