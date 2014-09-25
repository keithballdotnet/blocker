package blocks

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"fmt"
	"github.com/Inflatablewoman/blocks/crypto"
	"github.com/Inflatablewoman/blocks/hash2"
	"io"
	"os"
)

// Block is that basic element of data that is to be stored in the database
type Block struct {
	ID   string
	Hash []byte
	Data []byte
}

// This is a form used to link the File to the Block without needing to load the full data from the database
type FileBlock struct {
	ID            string
	BlockPosition int
	Hash          []byte
}

// File is a representation of a blocks together to form a file
type BlockedFile struct {
	ID      string
	Length  int64
	Version int
	Blocks  []FileBlock
}

// 4Mb block size
//const BlockSize int64 := 4194304

// Will start with small 30kb chunks to start with
const BlockSize int64 = 30720

// Create a new file.
// Expects a filename.  Returns any error or the ID of the new file
func BlockFile(sourceFilepath string, blockedFileID string) (error, BlockedFile) {

	// open the file and read the contents
	sourceFile, err := os.Open(sourceFilepath)
	if err != nil {
		return err, BlockedFile{}
	}
	defer sourceFile.Close()

	// Read in blocks of data
	data := make([]byte, BlockSize)

	fileblocks := make([]FileBlock, 0)

	// Create peristent store for blocks
	blockRepository, err := NewBlockRepository()
	if err != nil {
		return err, BlockedFile{}
	}

	blockedFileRepository, err := NewBlockedFileRepository()
	if err != nil {
		return err, BlockedFile{}
	}

	// Initial version
	blockedFileVersion := 1

	// var existingBlockFile BlockedFile
	existingBlockFile, err := blockedFileRepository.GetBlockedFile(blockedFileID)
	if err == nil && existingBlockFile != nil {
		blockedFileVersion = existingBlockFile.Version + 1
	}

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
		hash := hash2.ComputeSha256Checksum(data[:count])

		blockId := ""
		// If we have an existing version of this document.  Check to see if this block is the same.
		if existingBlockFile != nil {
			// Detect if block is same as existing block
			for _, existingFileBlock := range existingBlockFile.Blocks {
				if existingFileBlock.BlockPosition == blockCount && hash2.CompareChecksums(existingFileBlock.Hash, hash) {
					// We have a match for the block.  Set the blockId.  No need to store to repository
					blockId = existingFileBlock.ID
					fmt.Println("Reusing block:", blockId)
				}
			}
		}

		if blockId == "" {
			// Encrypt the data
			encryptedData, err := crypto.AesCfbEncrypt(data[:count])
			if err != nil {
				return err, BlockedFile{}
			}

			// Create our file structure
			block := Block{uuid.New(), hash, encryptedData}

			// Commit block to repository
			blockRepository.SaveBlock(block)
			blockId = block.ID

			fmt.Println("Created block:", blockId)
		}

		fileblock := FileBlock{blockId, blockCount, hash}

		// Add the file block to the list of blocks
		fileblocks = append(fileblocks, fileblock)

		fmt.Printf("Block #%d - ID %v read %d bytes\n", blockCount, blockId, count)
	}

	blockedFile := BlockedFile{uuid.New(), fileLength, blockedFileVersion, fileblocks}

	blockedFileRepository.SaveBlockedFile(blockedFile)

	return nil, blockedFile
}

func UnblockFile(blockFileID string, targetFilePath string) error {

	// Get the repository for the blockedFiles
	blockedFileRepository, err := NewBlockedFileRepository()
	if err != nil {
		return err
	}

	// Get peristent store for blocks
	blockRepository, err := NewBlockRepository()
	if err != nil {
		return err
	}

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

	fmt.Println("Got file: " + blockedFile.ID)

	var offSet int64 = 0
	for i, fileBlock := range blockedFile.Blocks {

		fmt.Printf("Got block #%d with ID %v\n", i+1, fileBlock.ID)

		block, err := blockRepository.GetBlock(fileBlock.ID)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}

		// Decrypt the data
		decryptedData, err := crypto.AesCfbDecrypt(block.Data)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}

		// Validate the hash
		if !hash2.ValidateSha256Checksum(decryptedData, block.Hash) {
			fmt.Println("Invalid block hash")
			return errors.New("Invalid block hash")
		}

		// Write out this block to the file
		bytesWritten, err := outFile.WriteAt(decryptedData, offSet)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}

		// Move offset
		offSet += int64(bytesWritten)

		fmt.Printf("Wrote %d bytes to file moving to offset %d\n", bytesWritten, offSet)
	}

	return nil
}
