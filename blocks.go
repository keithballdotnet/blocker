package blocks

import (
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/snappy-go/snappy"
	"fmt"
	"github.com/Inflatablewoman/blocks/crypto"
	"github.com/Inflatablewoman/blocks/hash2"
	"io"
	"os"
	"time"
)

// Block is that basic element of data that is to be stored in the database
type Block struct {
	Hash string `json:"hash"`
	Data []byte `json:"data"`
}

// This is a form used to link the File to the Block without needing to load the full data from the database
type FileBlock struct {
	BlockPosition int    `json:"position"`
	Hash          string `json:"hash"`
}

// File is a representation of a blocks together to form a file
type BlockedFile struct {
	ID        string      `json:"id"`
	Length    int64       `json:"length"`
	Created   time.Time   `json:"time"`
	BlockList []FileBlock `json:"blocks"`
}

// 4Mb block size
const BlockSize4Mb int64 = 4194304

// 1Mb block size
const BlockSize1Mb int64 = 1048576

// Will start with small 30kb chunks to start with
const BlockSize30Kb int64 = 30720

// 100kb block size
const BlockSize100Kb int64 = 102400

// Set default blocksize to 4Mb
var BlockSize int64 = BlockSize4Mb

// Compression is on by default
var UseCompression bool = true

// Use Encryption
var UseEncryption bool = true

// Repository for blockedFiles
var blockedFileRepository BlockedFileRepository

// Repository for blocks
var blockRepository BlockRepository

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
func BlockFile(sourceFilepath string) (error, BlockedFile) {

	// open the file and read the contents
	sourceFile, err := os.Open(sourceFilepath)
	if err != nil {
		return err, BlockedFile{}
	}
	defer sourceFile.Close()

	// Read in blocks of data
	data := make([]byte, BlockSize)

	fileblocks := make([]FileBlock, 0)

	var blockCount int
	var fileLength int64
	// Keep reading blocks of data from the file until we have read less than the BlockSize
	for count, err := sourceFile.Read(data); err == nil; count, err = sourceFile.Read(data) {
		blockCount++
		fileLength += int64(count)

		if err != nil && err != io.EOF {
			return err, BlockedFile{}
		}

		// Calculate the hash of the block
		hash := hash2.GetSha256HashString(data[:count])

		// Do we already have this block stored?
		blockExists, err := blockRepository.CheckBlockExists(hash)
		if err != nil {
			return err, BlockedFile{}
		}

		if !blockExists {

			storeData := data[:count]

			if UseCompression {
				// Compress the data
				storeData, err = snappy.Encode(nil, storeData)
				if err != nil {
					return err, BlockedFile{}
				}
			}

			if UseEncryption {
				// Encrypt the data
				storeData, err = crypto.AesCfbEncrypt(storeData)
				if err != nil {
					return err, BlockedFile{}
				}
			}

			// Create our file structure
			block := Block{hash, storeData}

			// Commit block to repository
			blockRepository.SaveBlock(block)
		}

		fileblock := FileBlock{blockCount, hash}

		// Add the file block to the list of blocks
		fileblocks = append(fileblocks, fileblock)
	}

	blockedFile := BlockedFile{uuid.New(), fileLength, time.Now(), fileblocks}

	blockedFileRepository.SaveBlockedFile(blockedFile)

	return nil, blockedFile
}

func UnblockFile(blockFileID string, targetFilePath string) error {

	// Get the blocked file from the repository
	blockedFile, err := blockedFileRepository.GetBlockedFile(blockFileID)
	if err != nil {
		return err
	}

	outFile, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var offSet int64 = 0
	for _, fileBlock := range blockedFile.BlockList {

		// fmt.Printf("Got block #%d with ID %v\n", i+1, fileBlock.ID)

		block, err := blockRepository.GetBlock(fileBlock.Hash)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}

		storeData := block.Data

		if UseEncryption {
			// Decrypt the data
			storeData, err = crypto.AesCfbDecrypt(storeData)
			if err != nil {
				fmt.Println("Error: " + err.Error())
				return err
			}
		}

		if UseCompression {
			// Uncompress the data
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
}
