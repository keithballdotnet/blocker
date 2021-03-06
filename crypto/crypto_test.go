package crypto

import (
	"bytes"
	"fmt"
	. "github.com/keithballdotnet/blocker/gocheck2"
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

func (s *CryptoSuite) TestAesGCMCrypto(c *C) {

	encryptString := "I once had a girl, or should I say, she once had me."

	bytesToEncrypt := []byte(encryptString)

	fmt.Println("GCM bytes to encrypt: " + string(bytesToEncrypt))

	aesKey := GenerateAesSecret()

	encryptedBytes, err := AesGCMEncrypt(bytesToEncrypt, aesKey)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("GCM encrypted bytes: " + string(encryptedBytes))

	unencryptedBytes, err := AesGCMDecrypt(encryptedBytes, aesKey)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("GCM Unencrypted bytes: " + string(unencryptedBytes))

	c.Assert(bytes.Equal(bytesToEncrypt, unencryptedBytes), IsTrue)
}

func (s *CryptoSuite) TestAesCrypto(c *C) {

	encryptString := "a very very very very secret pot"

	bytesToEncrypt := []byte(encryptString)

	fmt.Println("bytes to encrypt: " + string(bytesToEncrypt))

	encryptedBytes, err := AesCfbEncrypt(bytesToEncrypt, "testhash")

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("encrypted bytes: " + string(encryptedBytes))

	unencryptedBytes, err := AesCfbDecrypt(encryptedBytes, "testhash")

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

func (s *CryptoSuite) TestHMACKey(c *C) {

	expectedHmac := "RvPtP0QB7iIun1ehwheD4YUo7+fYfw7/ywl+HsC5Ddk="

	// The secret key
	secretKey := "e7yflbeeid26rredmwtbiyzxijzak6altcnrsi4yol2f5sexbgdwevlpgosfoeyy"
	method := "COPY"
	//date := time.Now().UTC().Format(time.RFC1123) // UTC time
	//fmt.Printf("Now: %s", date)
	date := "Wed, 28 Jan 2015 10:42:13 UTC"
	resource := "/api/v1/blocker/6f90d707-3b6a-4321-b32c-3c1d37915c1b"

	// Create auth request key
	authRequestKey := fmt.Sprintf("%s\n%s\n%s", method, date, resource)

	hmac := GetHmac256(authRequestKey, secretKey)

	// Test positive.
	c.Assert(expectedHmac == hmac, IsTrue, Commentf("HMAC wrong: %v Got Key %s", expectedHmac, hmac))

	// Test negative.  (Resource and Data in wrong order)
	authRequestKey = fmt.Sprintf("%s\n%s\n%s", method, resource, date)

	hmac = GetHmac256(authRequestKey, secretKey)

	// Test positive.
	c.Assert(expectedHmac != hmac, IsTrue, Commentf("HMAC should be differnt: %v Got Key %s", expectedHmac, hmac))
}
