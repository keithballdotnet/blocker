package blocks

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/snappy-go/snappy"
	"github.com/Inflatablewoman/blocker/crypto"
	"github.com/Inflatablewoman/blocker/hash2"
)

// This is a form used to link the File to the Block without needing to load the full data from the database
type Block struct {
	BlockPosition int    `json:"position"`
	Hash          string `json:"hash"`
}

// File is a representation of a blocks together to form a file
type BlockedFile struct {
	ID        string  `json:"id"`
	FileHash  string  `json:"fileHash"`
	Length    int64   `json:"length"`
	BlockList []Block `json:"blocks"`
}

// BlockInfo is used to maintain information about file blocks
type BlockInfo struct {
	Hash      string    `json:"hash"`
	StoreID   string    `json:"storeid"`
	UseCount  int64     `json:"usecount"`
	Created   time.Time `json:"created"`
	LastUsage time.Time `json:"lastUsed"`
}

// 4Mb block size
const BlockSize4Mb int64 = 4194304

// 1Mb block size
const BlockSize1Mb int64 = 1048576

// 30kb block size
const BlockSize30Kb int64 = 30720

// 100kb block size
const BlockSize100Kb int64 = 102400

// Set default blocksize to 4Mb
var BlockSize int64 = BlockSize4Mb

// Compression is on by default
var UseCompression bool = true

// Use Encryption is on by default
var UseEncryption bool = true

// Repository for blockedFiles
var BlockedFileStore BlockedFileRepository

// Repository for blocks
var BlockStore BlockRepository

// fileBlockInfoRepository for FileBlockInfo objects
var BlockInfoStore BlockInfoRepository

// StorageProviderName is the name of the selected storage provider
var StorageProviderName string

// CryptoProvider is the configured crypto provider to use for encrypting the data at rest
var CryptoProvider crypto.CryptoProvider

// CryptoProviderName is the name of the crypto provider
var CryptoProviderName string

// Set up repositories in the init to keep connections alive
func SetUpRepositories() {
	var err error
	// Create persistent store for BlockedFiles
	BlockedFileStore, err = NewBlockedFileRepository()
	if err != nil {
		panic(err)
	}

	// Create persistent store for FileBlockInfo
	BlockInfoStore, err = NewCouchbaseBlockInfoRepository()
	if err != nil {
		panic(err)
	}

	// Load the storage provider
	switch StorageProviderName {
	case "nfs":
		BlockStore, err = NewDiskBlockRepository()
	case "azure":
		BlockStore, err = NewAzureBlockRepository()
	case "cb":
		BlockStore, err = NewCouchBaseBlockRepository()
	case "s3":
		BlockStore, err = NewS3BlockRepository()
	default:
		// Default to storing to disk...
		BlockStore, err = NewDiskBlockRepository()
	}

	if err != nil {
		panic(err)
	}

	// Load the storage provider
	switch CryptoProviderName {
	case "aws":
		CryptoProvider, err = crypto.NewAwsCryptoProvider()
	case "openpgp":
		CryptoProvider, err = crypto.NewOpenPGPCryptoProvider()
	default:
		// Default to openpgp
		CryptoProvider, err = crypto.NewOpenPGPCryptoProvider()
	}

	if err != nil {
		panic(err)
	}
}

// Create a new file.
// Expects a filename.  Returns any error or the created BlockedFile
func BlockFile(sourceFilepath string) (BlockedFile, error) {

	// open the file and read the contents
	sourceFile, err := os.Open(sourceFilepath)
	if err != nil {
		return BlockedFile{}, err
	}
	defer sourceFile.Close()

	// Get blocked file (function used for testing so always same here)
	blockedFile, err := BlockBuffer(sourceFile)
	if err != nil {
		return BlockedFile{}, err
	}

	return blockedFile, nil
}

// Block a source into a file
func BlockBuffer(source io.Reader) (BlockedFile, error) {

	// Set up seeker
	readSeeker, _ := source.(io.ReadSeeker)

	// Create file hash
	fileHash := hash2.GetSha256HashStringFromStream(readSeeker)

	// Go back to start of stream for blocking
	readSeeker.Seek(0, 0)

	// Set the BlockSize
	data := make([]byte, BlockSize)

	fileblocks := make([]Block, 0)

	var blockCount int
	var fileLength int64

	// Keep reading blocks of data from the file until we have read less than the BlockSize
	for count, err := readSeeker.Read(data); err == nil; count, err = readSeeker.Read(data) {
		blockCount++
		fileLength += int64(count)

		if err != nil && err != io.EOF {
			return BlockedFile{}, err
		}

		// Calculate the hash of the block
		hash := hash2.GetSha256HashString(data[:count])

		// Get FileBlockInfo (if any)
		blockExists := false
		fileBlockInfo, err := BlockInfoStore.GetBlockInfo(hash)
		if err == nil {
			blockExists = true
		}

		// Get the time...
		now := time.Now().UTC()

		if !blockExists {
			storeData := data[:count]

			// Compress the data
			if UseCompression {
				storeData, err = snappy.Encode(nil, storeData)
				if err != nil {
					return BlockedFile{}, err
				}
			}

			// Encrypt the data
			if UseEncryption {
				storeData, err = CryptoProvider.Encrypt(storeData)
				if err != nil {
					return BlockedFile{}, err
				}
			}

			// Get a 50byte secret to store the file under
			storeID := strings.ToLower(crypto.RandomSecret(40))

			storeSize := len(storeData)
			log.Printf("Saving Block: %v Block: %v Store: %v (%.2f%%) StoreID: %v", hash, count, storeSize, ((float64(storeSize) / float64(count)) * 100), storeID)

			// Commit block to repository
			err = BlockStore.SaveBlock(storeData, storeID)
			if err != nil {
				return BlockedFile{}, err
			}

			// Save BlockInfo for hash
			err = BlockInfoStore.SaveBlockInfo(BlockInfo{Hash: hash, StoreID: storeID, UseCount: 1, Created: now, LastUsage: now})
		} else {
			// Register that we have been used again in another file
			fileBlockInfo.LastUsage = now
			fileBlockInfo.UseCount = fileBlockInfo.UseCount + 1
			err = BlockInfoStore.SaveBlockInfo(*fileBlockInfo)
		}

		fileblock := Block{blockCount, hash}

		// Add the file block to the list of blocks
		fileblocks = append(fileblocks, fileblock)
	}

	blockedFile := BlockedFile{uuid.New(), fileHash, fileLength, fileblocks}

	err := BlockedFileStore.SaveBlockedFile(blockedFile)

	return blockedFile, err
}

// DeleteBlockFile -  Deletes a BlockedFile and any unused FileBlocks
func DeleteBlockedFile(blockFileID string) error {
	// Get the blocked file from the repository
	blockedFile, err := BlockedFileStore.GetBlockedFile(blockFileID)
	if err != nil {
		return err
	}

	for _, fileBlock := range blockedFile.BlockList {
		// Store in the FileBlockInfo that we have been used...
		blockInfo, err := BlockInfoStore.GetBlockInfo(fileBlock.Hash)
		if err == nil {
			blockInfo.UseCount = blockInfo.UseCount - 1

			// Is the file block in use anymore?
			if blockInfo.UseCount < 1 {

				log.Printf("Deleting Hash: %v StoreID: %v", fileBlock.Hash, blockInfo.StoreID)

				// Delete from storage provider
				err = BlockStore.DeleteBlock(blockInfo.StoreID)
				if err != nil {
					return err
				}

				// Delete last instance of FileBlockInfo
				err = BlockInfoStore.DeleteBlockInfo(fileBlock.Hash)
				if err != nil {
					return err
				}

			} else {
				// Save that we are using the block one less time.
				BlockInfoStore.SaveBlockInfo(*blockInfo)
			}

		}
	}

	// Remove blocked file entry
	BlockedFileStore.DeleteBlockedFile(blockedFile.ID)

	return nil
}

// CopyBlockedFile -  Copy a blocked file and return the new BlockedFile
func CopyBlockedFile(blockFileID string) (BlockedFile, error) {
	// Get the blocked file from the repository
	blockedFile, err := BlockedFileStore.GetBlockedFile(blockFileID)
	if err != nil {
		return BlockedFile{}, err
	}

	// Create a copy of the BlockedFile and give it a new ID
	blockedFileCopy := *(blockedFile)
	blockedFileCopy.ID = uuid.New()
	BlockedFileStore.SaveBlockedFile(blockedFileCopy)

	// Update the FileBlockInfo for all the FileBlocks to maintain the use count...
	for _, fileBlock := range blockedFile.BlockList {
		// Store in the FileBlockInfo that we have been used...
		blockInfo, err := BlockInfoStore.GetBlockInfo(fileBlock.Hash)
		if err == nil {
			blockInfo.LastUsage = time.Now().UTC()
			blockInfo.UseCount = blockInfo.UseCount + 1
			// Save that we are using the block one more time.
			BlockInfoStore.SaveBlockInfo(*blockInfo)
		}
	}

	// Return the new copy
	return blockedFileCopy, nil
}

// Unblock a file to a buffer stream
func UnblockFileToBuffer(blockFileID string) (bytes.Buffer, error) {

	// Data to return
	var buffer bytes.Buffer

	// Get the blocked file from the repository
	blockedFile, err := BlockedFileStore.GetBlockedFile(blockFileID)
	if err != nil {
		return buffer, err
	}

	for _, fileBlock := range blockedFile.BlockList {

		blockInfo, err := BlockInfoStore.GetBlockInfo(fileBlock.Hash)
		if err != nil {
			log.Println("Error: " + err.Error())
			return buffer, err
		}

		log.Printf("Getting Hash: %v StoreID: %v", fileBlock.Hash, blockInfo.StoreID)

		bytes, err := BlockStore.GetBlock(blockInfo.StoreID)
		if err != nil {
			log.Println("Error: " + err.Error())
			return buffer, err
		}

		storeData := bytes

		// Decrypt the data
		if UseEncryption {
			storeData, err = CryptoProvider.Decrypt(storeData)
			if err != nil {
				log.Println("Error: " + err.Error())
				return buffer, err
			}
		}

		// Uncompress the data
		if UseCompression {
			storeData, err = snappy.Decode(nil, storeData)
			if err != nil {
				return buffer, err
			}
		}

		// Store in the FileBlockInfo that we have been used...
		blockInfo.LastUsage = time.Now().UTC()
		BlockInfoStore.SaveBlockInfo(*blockInfo)

		// Write data to buffer
		buffer.Write(storeData)
	}

	return buffer, nil
}

// Takes a file ID.  Unblocks the files from the underlying system and then writes the file to the target file path
func UnblockFile(blockFileID string, targetFilePath string) error {

	buffer, err := UnblockFileToBuffer(blockFileID)
	if err != nil {
		return err
	}

	outFile, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = buffer.WriteTo(outFile)
	if err != nil {
		log.Println("Error: " + err.Error())
		return err
	}

	return nil
}
