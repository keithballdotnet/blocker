package crypto

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// GoKMSCryptoProvider is an implementation of encryption using GO KMS
type GoKMSCryptoProvider struct {
	// keyID identifies which KMS key should be used for encryption / decryption
	keyID string
	// client identifies
	cli JSONClient
}

// NewGoKMSCryptoProvider
func NewGoKMSCryptoProvider() (GoKMSCryptoProvider, error) {

	log.Println("Using GoKMSCryptoProvider for encryption...")

	authKey := os.Getenv("BLOCKER_GOKMS_AUTHKEY")
	baseUrl := os.Getenv("BLOCKER_GOKMS_URL")

	if authKey == "" || baseUrl == "" {
		panic("Enivronmental Variable: BLOCKER_GOKMS_AUTHKEY or BLOCKER_GOKMS_URL are empty!  You must set these values when using GO KMS key management!")
	}

	client := http.DefaultClient

	ignoreBadTls := os.Getenv("BLOCKER_GOKMS_IGNORE_BAD_TLS_CERT")
	if strings.ToUpper(ignoreBadTls) == "TRUE" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		log.Println("WARNING: Ignore bad TLS Certificates is set to TRUE!  Do not do this in production!")
	}

	log.Printf("GoKMSCryptoProvider using GO-KMS @ %v", baseUrl)

	jsonClient := JSONClient{Client: client, Endpoint: baseUrl, AuthKey: authKey}
	gokms := GoKMSCryptoProvider{cli: jsonClient}

	keyID := os.Getenv("BLOCKER_GOKMS_KEYID")
	if keyID == "" {
		keyID = gokms.getNewestKeyID()
	}

	if keyID == "" {
		panic("Unable to find a key ID to use for encryption. You must set these values when using amazon KMS key management!")
	}

	return gokms, nil
}

// KeyByCreated - Will sort the Keys by CreationDate
type KeyByCreated []kms.KeyMetadata

func (a KeyByCreated) Len() int      { return len(a) }
func (a KeyByCreated) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a KeyByCreated) Less(i, j int) bool {
	return a[i].CreationDate.After(a[j].CreationDate)
}

func (p GoKMSCryptoProvider) getNewestKeyID() string {
	// List the key available...
	keyRequest := ListKeysRequest{}

	listKeyResponse := &ListKeysResponse{}
	err := p.cli.Do("POST", "/api/v1/go-kms/listkeys", &keyRequest, listKeyResponse)
	if err != nil {
		log.Printf("Unable to list keys: %v", err)
		return ""
	}

	// Make sure we pick the newest key for encryption...
	sort.Sort(KeyByCreated(listKeyResponse.KeyMetadata))

	for _, key := range listKeyResponse.KeyMetadata {
		log.Printf("Got key: %v %v", key.KeyID, key.Description)

		// Return the first key id..
		return key.KeyID
	}

	createKeyRequest := CreateKeyRequest{Description: "Blocker AES Encrypt/Decrypt Key"}

	createKeyResponse := &CreateKeyResponse{}
	err = p.cli.Do("POST", "/api/v1/go-kms/createkey", &createKeyRequest, createKeyResponse)
	if err != nil {
		log.Printf("Unable to create key: %v", err)
		return ""
	}

	return createKeyResponse.KeyMetadata.KeyID
}

// Encrypt will encrypt the passed data using a GO KMS key
func (p GoKMSCryptoProvider) Encrypt(data []byte) ([]byte, error) {

	// Request a new AES256 key from AWS KMS using the selected key
	generateKeyRequest := GenerateDataKeyRequest{KeyID: p.keyID}

	generateKeyResponse := &GenerateDataKeyResponse{}
	err := p.cli.Do("POST", "/api/v1/go-kms/generatedatakey", &generateKeyRequest, generateKeyResponse)
	if err != nil {
		log.Printf("Unable to get a data key: %v", err)
		return nil, err
	}

	// Encrypt data using AWS obtained key
	encryptedData, err := AesGCMEncrypt([]byte(data), generateKeyResponse.Plaintext)
	if err != nil {
		log.Printf("Unable to encrypt: %v", err)
		return nil, err
	}

	// Let's envelope the data
	var buffer bytes.Buffer
	// First write key
	buffer.Write(generateKeyResponse.CiphertextBlob)
	// Then write encrypted data
	buffer.Write(encryptedData)

	return buffer.Bytes(), nil
}

// Decrypt will decrypt the passed data using a GO KMS key
func (p GoKMSCryptoProvider) Decrypt(data []byte) ([]byte, error) {

	// Unpack envelope.

	// Get key from data
	keyPackageLength := 124
	keyPackage := make([]byte, keyPackageLength)

	// Now lets extra they key from the envelope
	envelopReader := bytes.NewReader(data)
	readCount, err := envelopReader.Read(keyPackage)
	if err != nil || readCount != keyPackageLength {
		log.Printf("Unable to get key from envelope: %v", err)
		return nil, err
	}

	// Read everything that is left...
	expectedDataLength := len(data) - keyPackageLength
	dataPackage := make([]byte, expectedDataLength)
	io.ReadFull(envelopReader, dataPackage)
	if err != nil {
		log.Printf("Unable to get data from envelope: %v", err)
		return nil, err
	}

	// Ask GO KMS to decrypt the key
	decryptRequest := DecryptRequest{CiphertextBlob: keyPackage}
	decryptResponse := &DecryptResponse{}
	err = p.cli.Do("POST", "/api/v1/go-kms/decrypt", &decryptRequest, decryptResponse)
	if err != nil {
		log.Printf("Unable to decrypt key package: %v", err)
		return nil, err
	}

	// Decrypt the datapackge with the unencrypted key
	decryptedData, err := AesGCMDecrypt(dataPackage, decryptResponse.Plaintext)
	if err != nil {
		log.Printf("Unable to decrypt data package: %v", err)
		return nil, err
	}

	return decryptedData, nil
}

// JSONClient is the underlying client for JSON APIs.
type JSONClient struct {
	Client   *http.Client
	Endpoint string
	// authKey is the key used for authentication
	AuthKey string
}

// Do sends an HTTP request and returns an HTTP response, following policy
// (e.g. redirects, cookies, auth) as configured on the client.
func (c *JSONClient) Do(method, uri string, req, resp interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(method, c.Endpoint+uri, bytes.NewReader(b))
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "Blocker")
	request.Header.Set("Content-Type", "application/json")

	request = c.SetAuth(request, method, uri)

	response, err := c.Client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return errors.New(fmt.Sprintf("KMSError StatusCode: %v Error: %v", response.StatusCode, string(bodyBytes)))
	}

	if resp != nil {
		return json.NewDecoder(response.Body).Decode(resp)
	}
	return nil
}

// SetAuth will set kms auth headers
func (c *JSONClient) SetAuth(request *http.Request, method string, resource string) *http.Request {

	date := time.Now().UTC().Format(time.RFC1123) // UTC time
	request.Header.Add("x-kms-date", date)

	authRequestKey := fmt.Sprintf("%s\n%s\n%s", method, date, resource)

	hmac := GetHmac256(authRequestKey, c.AuthKey)

	//fmt.Printf("SharedKey: %s HMAC: %s RequestKey: \n%s\n", SharedKey, hmac, authRequestKey)

	request.Header.Add("Authorization", hmac)

	return request
}

/* KMS JSON Structs */

// KeyMetadata is the associated meta data of any key
type KeyMetadata struct {
	KeyID        string    `json:"KeyId"`
	CreationDate time.Time `json:"CreationDate"`
	Description  string    `json:"Description"`
	Enabled      bool      `json:"Enabled"`
}

/* KMS Request / Response Structs */

// ReEncryptRequest
type ReEncryptRequest struct {
	CiphertextBlob   []byte `json:"CiphertextBlob"`
	DestinationKeyID string `json:"DestinationKeyId"`
}

// ReEncryptResponse
type ReEncryptResponse struct {
	CiphertextBlob []byte `json:"CiphertextBlob"`
	KeyID          string `json:"KeyID"`
	SourceKeyID    string `json:"SourceKeyID"`
}

// CreateKeyRequest
type CreateKeyRequest struct {
	Description string `json:"Description,omitempty"`
}

// CreateKeyResponse
type CreateKeyResponse struct {
	KeyMetadata KeyMetadata `json:"KeyMetadata"`
}

// listKeysHandler
type ListKeysRequest struct {
}

// ListKeysResponse
type ListKeysResponse struct {
	KeyMetadata []KeyMetadata `json:"KeyMetadata"`
}

// EnableKeyRequest
type EnableKeyRequest struct {
	KeyID string `json:"KeyID"`
}

// EnableKeyResponse
type EnableKeyResponse struct {
	KeyMetadata KeyMetadata `json:"KeyMetadata"`
}

// DisableKeyRequest
type DisableKeyRequest struct {
	KeyID string `json:"KeyID"`
}

// DisableKeyResponse
type DisableKeyResponse struct {
	KeyMetadata KeyMetadata `json:"KeyMetadata"`
}

// GenerateDataKeyRequest
type GenerateDataKeyRequest struct {
	KeyID string `json:"KeyID"`
}

// GenerateDataKeyResponse
type GenerateDataKeyResponse struct {
	Plaintext      []byte `json:"Plaintext"`
	CiphertextBlob []byte `json:"CiphertextBlob"`
}

// EncryptRequest
type EncryptRequest struct {
	KeyID     string `json:"KeyID"`
	Plaintext []byte `json:"Plaintext"`
}

// EncryptResponse
type EncryptResponse struct {
	CiphertextBlob []byte `json:"CiphertextBlob"`
}

// DecryptRequest
type DecryptRequest struct {
	CiphertextBlob []byte `json:"CiphertextBlob"`
}

// DecryptResponse
type DecryptResponse struct {
	Plaintext []byte `json:"Plaintext"`
}
