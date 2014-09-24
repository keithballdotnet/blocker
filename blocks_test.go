package blocks

import (
	"fmt"
	. "github.com/Inflatablewoman/blocks/gocheck2"
	. "gopkg.in/check.v1"
	"os"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type BlockSuite struct {
}

var _ = Suite(&BlockSuite{})

func (s *BlockSuite) TestCreateFile(c *C) {

	// NOTE:  Change this path
	const inputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest.txt"
	const outputFile = "/tmp/tempest.txt"

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)

	// Block the file
	err, blockFile := CreateFile(inputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// Check we read the full file size
	c.Assert(blockFile.Length == inputFileInfo.Size(), IsTrue)

	// Make sure the item returned some blocks
	c.Assert(len(blockFile.Blocks) > 0, IsTrue)

	// We have the file
	fmt.Println("Created new file: ", blockFile)

	// Clean up any old file
	os.Remove(outputFile)

	// Get the file and create a copy to the output
	err = GetFile(blockFile.ID, outputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(outputFile)

	fmt.Printf("Input filesize: %v\n", inputFileInfo.Size())
	fmt.Printf("Output filesize: %v\n", outputFileInfo.Size())

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == inputFileInfo.Size(), IsTrue)

}
