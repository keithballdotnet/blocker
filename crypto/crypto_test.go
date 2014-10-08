package crypto

import (
	"bytes"
	"fmt"
	. "github.com/Inflatablewoman/blocker/gocheck2"
	. "gopkg.in/check.v1"
	"os"
	// "path/filepath"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type CryptoSuite struct {
}

var _ = Suite(&CryptoSuite{})

func (s *CryptoSuite) TestAesCrypto(c *C) {

	encryptString := "a very very very very secret pot"

	bytesToEncrypt := []byte(encryptString)

	fmt.Println("bytes to encrypt: " + string(bytesToEncrypt))

	encryptedBytes, err := AesCfbEncrypt(bytesToEncrypt)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("encrypted bytes: " + string(encryptedBytes))

	unencryptedBytes, err := AesCfbDecrypt(encryptedBytes)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("Unencrypted bytes: " + string(unencryptedBytes))

	c.Assert(bytes.Equal(bytesToEncrypt, unencryptedBytes), IsTrue)
}

func (s *CryptoSuite) TestRsaCrypto(c *C) {

	encryptString := "a very very very very secret pot"

	bytesToEncrypt := []byte(encryptString)

	fmt.Println("bytes to encrypt: " + string(bytesToEncrypt))

	encryptedBytes, err := RsaEncrypt(bytesToEncrypt)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("encrypted bytes: " + string(encryptedBytes))

	unencryptedBytes, err := RsaDecrypt(encryptedBytes)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("Unencrypted bytes: " + string(unencryptedBytes))

	c.Assert(bytes.Equal(bytesToEncrypt, unencryptedBytes), IsTrue)
}

func (s *CryptoSuite) TestGenerateKey(c *C) {

	c.Skip("Not interesting")

	GenerateRsaKey()

	c.Assert(RsaEncryptionChipher.PublicKeyPath == CertifcatePath, IsTrue)
	c.Assert(RsaEncryptionChipher.PrivateKeyPath == KeyPath, IsTrue)

	certInfo, err := os.Stat(CertifcatePath)
	c.Assert(err == nil, IsTrue)
	c.Assert(certInfo.Size() > 0, IsTrue)

	keyInfo, err := os.Stat(KeyPath)
	c.Assert(err == nil, IsTrue)
	c.Assert(keyInfo.Size() > 0, IsTrue)
}
