package crypto

import (
	"bytes"
	"fmt"
	. "github.com/Inflatablewoman/blocker/gocheck2"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type CryptoPGPSuite struct {
}

var _ = Suite(&CryptoPGPSuite{})

// Path to the certificate
var publicPath = filepath.Join(os.TempDir(), "blocker", "public.pem")

// Path to the private key
var privatePath = filepath.Join(os.TempDir(), "blocker", "private.pem")

// Setup the REST testing suite
func (s *CryptoPGPSuite) SetUpSuite(c *C) {
	// Setup logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	ioutil.WriteFile(publicPath, []byte(publicKeyString), 0666)
	ioutil.WriteFile(privatePath, []byte(privateKeyString), 0666)
	os.Setenv("BLOCKER_PGP_PUBLICKEY", publicPath)
	os.Setenv("BLOCKER_PGP_PRIVATEKEY", privatePath)

	// Get the keys
	GetPGPKeyRings()
}

// Test down the suite
func (s *CryptoPGPSuite) TearDownSuite(c *C) {
	/*os.Remove(publicPath)
	os.Remove(privatePath)
	os.Setenv("BLOCKER_PGP_PUBLICKEY", "")
	os.Setenv("BLOCKER_PGP_PRIVATEKEY", "")*/
}

func (s *CryptoPGPSuite) TestPGPCrypto(c *C) {

	encryptString := "encrypting the string with pgp"

	bytesToEncrypt := []byte(encryptString)

	fmt.Println("bytes to encrypt: " + string(bytesToEncrypt))

	encryptedBytes, err := PGPEncrypt(bytesToEncrypt)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("encrypted bytes: " + string(encryptedBytes))

	unencryptedBytes, err := PGPDecrypt(encryptedBytes)

	if err != nil {
		fmt.Println("Got error: " + err.Error())
	}

	// No error
	c.Assert(err == nil, IsTrue)

	fmt.Println("Unencrypted bytes: " + string(unencryptedBytes))

	c.Assert(bytes.Equal(bytesToEncrypt, unencryptedBytes), IsTrue)
}

var publicKeyString = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

mQENBFTOoSkBCACumCLr1wooINzzmEs6/TjaolPOas03anH64fE/WdEj2cCPKp6d
6IyesHLnn8wxZfk07wC2CIy1e6YryDTFRiriWa+vgpSqRD/sEF76oKy3JsmNmSrA
WCI3QqRjL1e1DET5tyLK86DnOBqGqCLKNCIJmK9DCz0tMkC3xLNAkpmpSeHuCE6e
HUskawj13MnlgA77PZFQGaywTa/0J1YLaBNQxtmAraWeXAPW5JKia7l5H4dw+2IO
sDGWO3yvuTNxvkSQazCfhi1tpVSu8tGIweYEPjruV9vBdfVVsKo//v+2hFTGTrjC
kGj+NX2IqgXb5Lv9oOT19UmbU9Rr8gW8BwtRABEBAAG0NWJsb2NrZXIgKGJsb2Nr
ZXIga2V5KSA8a2VpdGguYmFsbC5ibG9ja2VyQGdpdGh1Yi5jb20+iQE5BBMBAgAj
BQJUzqEpAhsDBwsJCAcDAgEGFQgCCQoLBBYCAwECHgECF4AACgkQqYMASGoZ6MGh
BQf/X5oFbG6HJNYKaoYCmzvAu6ibaG3PW9KQYiMgoIKrbCOH8IQ2AWSmuHJIlBWS
dZ1EJizRCTC0pFCS7jy9aOoTIFgcDhhl48mA3K+UDTO95vyuFapJhNEDjyMQkHrX
bixj/Hhn/mb0WSkcrwsdnJh2d/xkfv958Bcyh72SyV36B4bmEupzV+vJpR1eYxnk
adiDn9luFGbLL6GsAAEpuGzr0pXZsKwXFyiPFaSLG5M2zuvjFKZ6Kt/nO7yc86rr
ndhSB8fwmSCTOe0kx/2jTWTj6HSw0DfyCLSLfGoeaJsayj4oybbksKB7zNZgUulr
coqzoLxN+zaCTSYBLSd9ideffbkBDQRUzqEpAQgAtZ1i3COwpJW7EhDL+lvlspZp
ifqAh3H8iRJlNYyitEJ7HWYJSL6Kw/42CHLPQARQK2qHvttZ+huxOUvYo9rN/h6Q
Pj+GYl9x0YuMMdOu3V/WpiZa2wuXHKcnxfuEPXMdxwFsGMtzF3gVzK+REWqAgDQA
isx14ngHJEnMycKRoGiL9yywhZaqJ7qK4M7J9gIcdj8fO3Tc3s8II2XZLbsrq69n
ui7/qmVKAlqBj1F+FT3y80N9YDjeSfeURl0NnXNR9pyG0F5QI8A5XA6TG/TEgsde
2lZpohCtq5PsbOmNKd1ppnCAWvFNJh7XX7lY9LZVqEUXRnddm3qf7WAzin7NzQAR
AQABiQEfBBgBAgAJBQJUzqEpAhsMAAoJEKmDAEhqGejBrgYH/3/7ipeJZyrKdFJ9
v51xmzMkKABXOB9nzQeiEowc5UDMTdaYQ7FU9d6tqEQKb05GPybDZCe0N+UVxYBI
ygUkyKiNKpGWnSwTNwrKgSYxJOh3jnxhP26moKMyIaZUSBDS3bBo03ZI7EThdl6w
88vSKm9T2wahu7sH23nGvFQ0nZU4MRmfVm9veCeuAvymVb8fVFWWPAJLDwLBt1Rm
DAR+pKke4PBz9aDp0YwNDK5TbCzj4BHwSufd+SH9HBDCZh7bG//zIQ52XlchEw+q
RKWWRc0hauFpnxe1BRui2ogaLP7EKatLDK88afQ+XMmsyLxUmghM9rwgt69cmh9z
fxLHKOo=
=6gVY
-----END PGP PUBLIC KEY BLOCK-----`

var privateKeyString = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v2

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
t69cmh9zfxLHKOo=
=kbne
-----END PGP PRIVATE KEY BLOCK-----`
