package crypto

import (
	"bytes"
	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/gen/kms"
	"io"
	"log"
	"os"
)

// AwsCryptoProvider is an implementation of encryption using AWS KMS
type AwsCryptoProvider struct {
	// cli is The KMS client
	cli *kms.KMS
	// keyID identifies which KMS key should be used for encryption / decryption
	keyID string
}

// NewAwsCryptoProvider
func NewAwsCryptoProvider() (AwsCryptoProvider, error) {

	log.Println("Using AwsCryptoProvider for encryption...")

	awskey := os.Getenv("BLOCKER_KMS_KEY")
	awssecret := os.Getenv("BLOCKER_KMS_SECRET")

	if awskey == "" || awssecret == "" {
		panic("Enivronmental Variable: BLOCKER_KMS_KEY or BLOCKER_KMS_SECRET are empty!  You must set these values when using amazon KMS key management!")
	}

	// Set up credentials...
	creds := aws.Creds(awskey, awssecret, "")

	awsregion := os.Getenv("BLOCKER_KMS_REGION")

	// Default to central region
	if awsregion == "" {
		awsregion = "eu-central-1"
	}

	// Connect to eu west
	cli := kms.New(creds, awsregion, nil)

	awskeyID := os.Getenv("BLOCKER_KMS_KEY_ID")
	if awskeyID == "" {
		awskeyID = getNewestKeyID(cli)
	}

	if awskeyID == "" {
		panic("Unable to find a key ID to use for encryption. You must set these values when using amazon KMS key management!")
	}

	log.Printf("AwsCryptoProvider using Key: %v", awskeyID)

	return AwsCryptoProvider{cli: cli, keyID: awskeyID}, nil
}

func getNewestKeyID(cli *kms.KMS) string {
	// List the key available...
	keyRequest := kms.ListKeysRequest{}
	listKeyResponse, err := cli.ListKeys(&keyRequest)
	if err != nil {
		log.Printf("Unable to list keys: %v", err)
		return ""
	}

	// TODO: Sort this list and select newest key that can do encryption...
	for _, key := range listKeyResponse.Keys {
		log.Printf("Got key: %v %v", *(key.KeyARN), *(key.KeyID))

		// Return the first key id..
		return *(key.KeyID)

		/*describeKeyRequest := kms.DescribeKeyRequest{KeyID: aws.String(*(key.KeyID))}
		describeKeyResponse, err := cli.DescribeKey(&describeKeyRequest)
		if err != nil {
			log.Printf("Unable to describe key: %v", err)
			return ""
		}

		log.Printf("KeyMetadata.ARN: %v", *(describeKeyResponse.KeyMetadata.ARN))
		log.Printf("KeyMetadata.AWSAccountID: %v", *(describeKeyResponse.KeyMetadata.AWSAccountID))
		log.Printf("KeyMetadata.CreationDate: %v", *(describeKeyResponse.KeyMetadata.CreationDate))
		log.Printf("KeyMetadata.Description: %v", *(describeKeyResponse.KeyMetadata.Description))
		log.Printf("KeyMetadata.Enabled: %v", *(describeKeyResponse.KeyMetadata.Enabled))
		log.Printf("KeyMetadata.KeyID: %v", *(describeKeyResponse.KeyMetadata.KeyID))
		log.Printf("KeyMetadata.KeyUsage: %v", *(describeKeyResponse.KeyMetadata.KeyUsage))*/
	}

	return ""
}

// Encrypt will encrypt the passed data using a AWS KMS key
func (p AwsCryptoProvider) Encrypt(data []byte) ([]byte, error) {

	// Request a new AES256 key from AWS KMS using the selected key
	generateKeyRequest := kms.GenerateDataKeyRequest{KeyID: aws.String(p.keyID), KeySpec: aws.String(kms.DataKeySpecAES256)}
	generateKeyResponse, err := p.cli.GenerateDataKey(&generateKeyRequest)
	if err != nil {
		log.Printf("Unable to get a data key: %v", err)
		return nil, err
	}

	// Encrypt data using AWS obtained key
	encryptedData, err := AesEncrypt([]byte(data), generateKeyResponse.Plaintext)
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

// Decrypt will decrypt the passed data using a AWS KMS key
func (p AwsCryptoProvider) Decrypt(data []byte) ([]byte, error) {

	// Unpack envelope.

	// Get key from data
	keyPackageLength := 204
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

	// Ask AWS KMS to decrypt the key
	decryptRequest := kms.DecryptRequest{CiphertextBlob: keyPackage}
	decryptResponse, err := p.cli.Decrypt(&decryptRequest)
	if err != nil {
		log.Printf("Unable to decrypt key package: %v", err)
		return nil, err
	}

	// Decrypt the datapackge with the unencrypted key
	decryptedData, err := AesDecrypt(dataPackage, decryptResponse.Plaintext)
	if err != nil {
		log.Printf("Unable to decrypt data package: %v", err)
		return nil, err
	}

	return decryptedData, nil
}
