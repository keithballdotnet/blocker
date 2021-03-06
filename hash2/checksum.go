package hash2

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// This returns the data's MD5 checksum.
//
// WARNING: Do NOT Use MD5 in security contexts (defending against
// intentional manipulations of data from untrusted sources);
// use only for checking data integrity against machine errors.
func ComputeMd5Checksum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

// This returns the data's Sha256 checksum
func ComputeSha256Checksum(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// This returns the data's Sha256 checksum as a string representation
func GetSha256HashString(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// Return a hash from a stream reader
func GetSha256HashStringFromStream(stream io.Reader) string {
	hash := sha256.New()
	io.Copy(hash, stream)
	return hex.EncodeToString(hash.Sum(nil))
}

// This returns true if the data matches the provided checksum.
func ValidateMd5Checksum(data []byte, sum []byte) bool {
	ourSum := ComputeMd5Checksum(data)
	return bytes.Equal(ourSum, sum)
}

// This returns true if the data matches the provided checksum.
func ValidateSha256Checksum(data []byte, sum []byte) bool {
	ourSum := ComputeSha256Checksum(data)
	return bytes.Equal(ourSum, sum)
}

// Compare two check sums.  Return true if they are they same
func CompareChecksums(first []byte, second []byte) bool {
	return bytes.Equal(first, second)
}
