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
	const changedInputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest_changed.txt"
	const outputFile = "/tmp/tempest.txt"

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)

	// Block the file
	err, blockFile := BlockFile(inputFile, "")

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// File is new should be version 1
	c.Assert(blockFile.Version == 1, IsTrue)

	// Check we read the full file size
	c.Assert(blockFile.Length == inputFileInfo.Size(), IsTrue)

	// Make sure the item returned some blocks
	c.Assert(len(blockFile.Blocks) > 0, IsTrue)

	// We have the file
	fmt.Println("Block file: ", blockFile)

	// Clean up any old file
	os.Remove(outputFile)

	// Get the file and create a copy to the output
	err = UnblockFile(blockFile.ID, outputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(outputFile)

	fmt.Printf("Input filesize: %v\n", inputFileInfo.Size())
	fmt.Printf("Output filesize: %v\n", outputFileInfo.Size())

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == inputFileInfo.Size(), IsTrue)

	// Block the file again.  New version should be created
	err, blockFile = BlockFile(inputFile, blockFile.ID)

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// File is new version so should be version 2
	c.Assert(blockFile.Version == 2, IsTrue)

	// We have the file
	fmt.Println("Block file: ", blockFile)

	// Block the file again.  New version should be created
	err, blockFile = BlockFile(changedInputFile, blockFile.ID)

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// File is new version so should be version 3
	c.Assert(blockFile.Version == 3, IsTrue)

	// We have the file
	fmt.Println("Block file: ", blockFile)
}
