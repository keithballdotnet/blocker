package blocks

import (
	"fmt"
	. "github.com/Inflatablewoman/blocks/gocheck2"
	. "gopkg.in/check.v1"
	"os"
	"testing"
	"time"
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
	const bibleInFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/kjv.txt"
	const bibleOutFile = "/tmp/kjv.txt"
	const changedInputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest_changed.txt"
	const changedAgainInputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest_changed_again.txt"
	const outputFile = "/tmp/tempest.txt"
	const changedOutputFile = "/tmp/tempest_changed.txt"

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)
	// Get some info about the file we are going test
	changedInputFileInfo, _ := os.Stat(changedInputFile)

	// Block the file
	start := time.Now()
	err, blockFile := BlockFile(inputFile, "")
	end := time.Now()

	fmt.Printf("Blocked Tempest took: %v\n", end.Sub(start))

	// Block the bigger
	start = time.Now()
	err, bibleBlockFile := BlockFile(bibleInFile, "")
	end = time.Now()

	fmt.Printf("Blocked King James Bible took: %v\n", end.Sub(start))

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
	start = time.Now()
	err = UnblockFile(blockFile.ID, outputFile)
	end = time.Now()

	fmt.Printf("Unblocked tempest took: %v\n", end.Sub(start))

	// Clean up any old file
	os.Remove(bibleOutFile)

	// Get the file and create a copy to the output
	start = time.Now()
	err = UnblockFile(bibleBlockFile.ID, bibleOutFile)
	end = time.Now()

	fmt.Printf("Unblocked King James Bible took: %v\n", end.Sub(start))

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

	// Clean up any old file
	os.Remove(changedOutputFile)

	// Get the file and create a copy to the output
	err = UnblockFile(blockFile.ID, changedOutputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Get some info about the file we are going test
	outputFileInfo, _ = os.Stat(changedOutputFile)

	fmt.Printf("Input (CHANGED) filesize: %v\n", changedInputFileInfo.Size())
	fmt.Printf("Output (CHANGED) filesize: %v\n", outputFileInfo.Size())

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == changedInputFileInfo.Size(), IsTrue)
}
