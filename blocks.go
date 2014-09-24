package blocks

import (
	"fmt"
	"os"
)

// Block is that basic element of data that is to be stored in the database
type Block struct {
	ID   string
	Hash []byte
	Data []byte
}

// File is a representation of a blocks together to form a file
type File struct {
	ID       string
	BlockIDs []string
}

// 4Mb block size
//const BlockSize int64 := 4194304

// Will start with small 30kb chunks to start with
const BlockSize int64 = 30720

// Create a new file.
// Expects a filename.  Returns any error or the ID of the new file
func CreateFile(sourceFilepath string) (error, string) {

	// open the file and read the contents
	sourceFile, err := os.Open(sourceFilepath)
	if err != nil {
		return err, ""
	}
	defer sourceFile.Close()

	// Read in blocks of data
	data := make([]byte, BlockSize)
	count, err := sourceFile.Read(data)
	if err != nil {
		return err, ""
	}

	fmt.Printf("read %d bytes: %q\n", count, data[:count])

	return nil, "I will make an ID"
}
