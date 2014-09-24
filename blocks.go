package blocks

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
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
	ID   string
	Hash []byte
}

// File is a representation of a blocks together to form a file
type BlockedFile struct {
	ID     string
	Blocks []FileBlock
}

// 4Mb block size
//const BlockSize int64 := 4194304

// Will start with small 30kb chunks to start with
const BlockSize int64 = 30720

// Create a new file.
// Expects a filename.  Returns any error or the ID of the new file
func CreateFile(sourceFilepath string) (error, BlockedFile) {

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

	var blockCount int
	// Keep reading blocks of data from the file until we have read less than the BlockSize
	for count, err := sourceFile.Read(data); err == nil; count, err = sourceFile.Read(data) {
		blockCount++

		if err != nil && err != io.EOF {
			return err, BlockedFile{}
		}

		// Calculate the hash of the block
		hash := hash2.ComputeSha256Checksum(data)

		// Create our file structure
		block := Block{uuid.New(), hash, data}

		blockRepository.SaveBlock(block)

		fileblock := FileBlock{block.ID, hash}

		// Add the file block to the list of blocks
		fileblocks = append(fileblocks, fileblock)

		fmt.Println("Created block:", block.ID)

		// , data[:count]

		fmt.Printf("Block #%d - ID %d read %d bytes with hash %v\n", blockCount, block.ID, count, hash)

	}

	blockedFile := BlockedFile{uuid.New(), fileblocks}

	return nil, blockedFile
}
