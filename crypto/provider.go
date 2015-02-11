package crypto

// CryptoProvider provides an interface for crypto provider solutions
type CryptoProvider interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}
