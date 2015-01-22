package blocks

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/snappy-go/snappy"
	"github.com/Inflatablewoman/blocker/crypto"
	"github.com/Inflatablewoman/blocker/hash2"
)

// This is a form used to link the File to the Block without needing to load the full data from the database
type FileBlock struct {
	BlockPosition int    `json:"position"`
	Hash          string `json:"hash"`
}

// File is a representation of a blocks together to form a file
type BlockedFile struct {
	ID        string      `json:"id"`
	FileHash  string      `json:"fileHash"`
	Length    int64       `json:"length"`
	BlockList []FileBlock `json:"blocks"`
}

// FileBlockInfo is used to maintain information about file blocks
type FileBlockInfo struct {
	Hash      string    `json:"hash"`
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
var blockedFileRepository BlockedFileRepository

// Repository for blocks
var blockRepository BlockRepository

// fileBlockInfoRepository for FileBlockInfo objects
var fileBlockInfoRepository FileBlockInfoRepository

// StorageProviderName is the name of the selected storage provider
var StorageProviderName string

// Set up repositories in the init to keep connections alive
func SetUpRepositories() {
	var err error
	// Create persistent store for BlockedFiles
	blockedFileRepository, err = NewBlockedFileRepository()
	if err != nil {
		panic(err)
	}

	// Create persistent store for FileBlockInfo
	fileBlockInfoRepository, err = NewCouchbaseFileBlockInfoRepository()
	if err != nil {
		panic(err)
	}

	// Load the storage provider
	switch StorageProviderName {
	case "nfs":
		blockRepository, err = NewDiskBlockRepository()
	case "azure":
		blockRepository, err = NewAzureBlockRepository()
	case "cb":
		blockRepository, err = NewCouchBaseBlockRepository()
	case "s3":
		panic("Not Implemented")
	default:
		// Default to storing to disk...
		blockRepository, err = NewDiskBlockRepository()
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

	fileblocks := make([]FileBlock, 0)

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
		fileBlockInfo, err := fileBlockInfoRepository.GetFileBlockInfo(hash)
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
				storeData, err = crypto.AesCfbEncrypt(storeData, hash)
				if err != nil {
					return BlockedFile{}, err
				}
			}

			// Commit block to repository
			blockRepository.SaveBlock(storeData, hash)

			// Save FileBlockInfo for hash

			fileBlockInfoRepository.SaveFileBlockInfo(FileBlockInfo{Hash: hash, UseCount: 1, Created: now, LastUsage: now})
		} else {
			// Register that we have been used again in another file
			fileBlockInfo.LastUsage = now
			fileBlockInfo.UseCount = fileBlockInfo.UseCount + 1
			fileBlockInfoRepository.SaveFileBlockInfo(*fileBlockInfo)
		}

		fileblock := FileBlock{blockCount, hash}

		// Add the file block to the list of blocks
		fileblocks = append(fileblocks, fileblock)
	}

	blockedFile := BlockedFile{uuid.New(), fileHash, fileLength, fileblocks}

	blockedFileRepository.SaveBlockedFile(blockedFile)

	return blockedFile, nil
}

// DeleteBlockFile -  Deletes a BlockedFile and any unused FileBlocks
func DeleteBlockedFile(blockFileID string) error {
	// Get the blocked file from the repository
	blockedFile, err := blockedFileRepository.GetBlockedFile(blockFileID)
	if err != nil {
		return err
	}

	for _, fileBlock := range blockedFile.BlockList {
		// Store in the FileBlockInfo that we have been used...
		fileBlockInfo, err := fileBlockInfoRepository.GetFileBlockInfo(fileBlock.Hash)
		if err == nil {
			fileBlockInfo.UseCount = fileBlockInfo.UseCount - 1
			log.Printf("Block: %s UseCount: %v", fileBlock.Hash, fileBlockInfo.UseCount)

			// Is the file block in use anymore?
			if fileBlockInfo.UseCount < 1 {

				log.Printf("Deleting Block: %s", fileBlock.Hash)

				// Delete from storage provider
				err = blockRepository.DeleteBlock(fileBlock.Hash)
				if err != nil {
					return err
				}

				// Delete last instance of FileBlockInfo
				err = fileBlockInfoRepository.DeleteFileBlockInfo(fileBlock.Hash)
				if err != nil {
					return err
				}

				// Remove the key
				crypto.DeleteAesSecret(fileBlock.Hash)
			} else {
				// Save that we are using the block one less time.
				fileBlockInfoRepository.SaveFileBlockInfo(*fileBlockInfo)
			}

		}

	}

	// Remove blocked file entry
	blockedFileRepository.DeleteBlockedFile(blockedFile.ID)

	return nil
}

// Unblock a file to a buffer stream
func UnblockFileToBuffer(blockFileID string) (bytes.Buffer, error) {

	// Data to return
	var buffer bytes.Buffer

	// Get the blocked file from the repository
	blockedFile, err := blockedFileRepository.GetBlockedFile(blockFileID)
	if err != nil {
		return buffer, err
	}

	for _, fileBlock := range blockedFile.BlockList {

		bytes, err := blockRepository.GetBlock(fileBlock.Hash)
		if err != nil {
			log.Println("Error: " + err.Error())
			return buffer, err
		}

		storeData := bytes

		// Decrypt the data
		if UseEncryption {
			storeData, err = crypto.AesCfbDecrypt(storeData, fileBlock.Hash)
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
		fileBlockInfo, err := fileBlockInfoRepository.GetFileBlockInfo(fileBlock.Hash)
		if err == nil {
			fileBlockInfo.LastUsage = time.Now().UTC()
			fileBlockInfoRepository.SaveFileBlockInfo(*fileBlockInfo)
		}

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
