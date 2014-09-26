package hash2

import (
	"encoding/hex"
	. "github.com/Inflatablewoman/blocks/gocheck2"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ChecksumSuite struct {
}

var _ = Suite(&ChecksumSuite{})

var input = []byte("it just works")
var md5Reference = []byte("qp\xbdZ\x9dSH\xe81\x10\x0fk\x81\xff\xda\xdc")

func (s *ChecksumSuite) TestHash(c *C) {
	ourMd5 := ComputeMd5Checksum(input)
	c.Assert(ourMd5, DeepEquals, md5Reference)

	c.Assert(ValidateMd5Checksum(input, md5Reference), IsTrue)
	c.Assert(ValidateMd5Checksum(input, ourMd5), IsTrue)

	// Fails...
	c.Assert(ValidateMd5Checksum(ourMd5, input), IsFalse)

	// No panic on empty
	c.Assert(ValidateMd5Checksum([]byte(""), []byte("")), IsFalse)

	ourSha256 := ComputeSha256Checksum(input)
	c.Assert(ValidateSha256Checksum(input, ourSha256), IsTrue)

	sha256AsString := GetSha256HashString(input)

	// Make sure not empty string
	c.Assert(sha256AsString != "", IsTrue)

	// Ensure is the same as out previous hash
	firstHashString := "sha256:" + hex.EncodeToString(ourSha256)
	c.Assert(firstHashString == sha256AsString, IsTrue)
}
