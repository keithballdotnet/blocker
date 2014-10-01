package blocks

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/snappy-go/snappy"
	"fmt"
	"github.com/Inflatablewoman/blocker/crypto"
	"github.com/Inflatablewoman/blocker/hash2"
	"io"
	"os"
	"path/filepath"
	"time"
)

// This is a form used to link the File to the Block without needing to load the full data from the database
type FileBlock struct {
	BlockPosition int    `json:"position"`
	Hash          string `json:"hash"`
}

// File is a representation of a blocks together to form a file
type BlockedFile struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	ContentType string      `json:"contentType"`
	Length      int64       `json:"length"`
	Created     time.Time   `json:"time"`
	BlockList   []FileBlock `json:"blocks"`
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

// Set up repositories in the init to keep connections alive
func init() {
	var err error
	// Create persistent store for BlockedFiles
	blockedFileRepository, err = NewBlockedFileRepository()
	if err != nil {
		panic(err)
	}

	// Create peristent store for blocks
	blockRepository, err = NewBlockRepository()
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

	filename := filepath.Base(sourceFilepath)

	// Get blocked file (function used for testing so always same here)
	blockedFile, err := BlockBuffer(sourceFile, filename, "plain/text")
	if err != nil {
		return BlockedFile{}, err
	}

	return blockedFile, nil
}

// Block a source into a file
func BlockBuffer(source io.Reader, filename string, fileType string) (BlockedFile, error) {

	// Set the BlockSize
	data := make([]byte, BlockSize)

	fileblocks := make([]FileBlock, 0)

	var blockCount int
	var fileLength int64

	// Keep reading blocks of data from the file until we have read less than the BlockSize
	for count, err := source.Read(data); err == nil; count, err = source.Read(data) {
		blockCount++
		fileLength += int64(count)

		if err != nil && err != io.EOF {
			return BlockedFile{}, err
		}

		// Calculate the hash of the block
		hash := hash2.GetSha256HashString(data[:count])

		// Do we already have this block stored?
		blockExists, err := blockRepository.CheckBlockExists(hash)
		if err != nil {
			return BlockedFile{}, err
		}

		fmt.Printf("count: %d\n", count)

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
				storeData, err = crypto.AesCfbEncrypt(storeData)
				if err != nil {
					return BlockedFile{}, err
				}
			}

			// Commit block to repository
			blockRepository.SaveBlock(storeData, hash)
		}

		fileblock := FileBlock{blockCount, hash}

		// Add the file block to the list of blocks
		fileblocks = append(fileblocks, fileblock)
	}

	blockedFile := BlockedFile{uuid.New(), filename, fileType, fileLength, time.Now(), fileblocks}

	blockedFileRepository.SaveBlockedFile(blockedFile)

	return blockedFile, nil
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
			fmt.Println("Error: " + err.Error())
			return buffer, err
		}

		storeData := bytes

		// Decrypt the data
		if UseEncryption {
			storeData, err = crypto.AesCfbDecrypt(storeData)
			if err != nil {
				fmt.Println("Error: " + err.Error())
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
		fmt.Println("Error: " + err.Error())
		return err
	}

	return nil
}

/* // Get the blocked file from the repository
	blockedFile, err := blockedFileRepository.GetBlockedFile(blockFileID)
	if err != nil {
		return err
	}


	var offSet int64 = 0
	for _, fileBlock := range blockedFile.BlockList {

		// fmt.Printf("Got block #%d with ID %v\n", i+1, fileBlock.ID)

		block, err := blockRepository.GetBlock(fileBlock.Hash)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}

		storeData := block.Data

		// Decrypt the data
		if UseEncryption {
			storeData, err = crypto.AesCfbDecrypt(storeData)
			if err != nil {
				fmt.Println("Error: " + err.Error())
				return err
			}
		}

		// Uncompress the data
		if UseCompression {

			storeData, err = snappy.Decode(nil, storeData)
			if err != nil {
				return err
			}
		}

		// Write out this block to the file
		bytesWritten, err := outFile.WriteAt(storeData, offSet)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}

		// Move offset
		offSet += int64(bytesWritten)

		// fmt.Printf("Wrote %d bytes to file moving to offset %d\n", bytesWritten, offSet)
	}

	return nil
}*/
