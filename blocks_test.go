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

const inputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest.txt"
const bibleInFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/kjv.txt"
const bibleOutFile = "/tmp/kjv.txt"
const changedInputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest_changed.txt"
const changedAgainInputFile = "/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest_changed_again.txt"
const outputFile = "/tmp/tempest.txt"
const changedOutputFile = "/tmp/tempest_changed.txt"

func (s *BlockSuite) Test1MbKingJamesBible(c *C) {

	BlockSize = BlockSize1Mb

	// Block the bigger
	start := time.Now()
	err, bibleBlockFile := BlockFile(bibleInFile)
	end := time.Now()

	fmt.Printf("Blocked King James Bible took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue)

	// Clean up any old file
	os.Remove(bibleOutFile)

	// Get the file and create a copy to the output
	start = time.Now()
	err = UnblockFile(bibleBlockFile.ID, bibleOutFile)
	end = time.Now()

	fmt.Printf("Unblocked King James Bible took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue)
}

func (s *BlockSuite) Test30KbTempest(c *C) {
	BlockSize = BlockSize30Kb

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)

	// Block the file
	start := time.Now()
	err, blockFile := BlockFile(inputFile)
	end := time.Now()

	fmt.Printf("Blocked Tempest at 30Kb took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// Check it was created in the past
	c.Assert(blockFile.Created.Before(time.Now()), IsTrue)

	// Check we read the full file size
	c.Assert(blockFile.Length == inputFileInfo.Size(), IsTrue)

	// Make sure the item returned some blocks
	c.Assert(len(blockFile.BlockList) > 0, IsTrue)

	// Clean up any old file
	os.Remove(outputFile)

	// Get the file and create a copy to the output
	start = time.Now()
	err = UnblockFile(blockFile.ID, outputFile)
	end = time.Now()

	fmt.Printf("Unblocked Tempest took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue)

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(outputFile)

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == inputFileInfo.Size(), IsTrue)
}

func (s *BlockSuite) TestChangeTempest(c *C) {

	// Use 1Mb Blocks
	BlockSize = BlockSize1Mb

	// Get some info about the file we are going test
	changedInputFileInfo, _ := os.Stat(changedInputFile)

	err, blockFile := BlockFile(inputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	firstFileHash := blockFile.BlockList[0].Hash

	// Block the file again.
	err, blockFile = BlockFile(inputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// Check that block used in first block is the same
	c.Assert(firstFileHash == blockFile.BlockList[0].Hash, IsTrue)

	// Block the file again.  New version should be created
	err, blockFile = BlockFile(changedInputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// Clean up any old file
	os.Remove(changedOutputFile)

	// Get the file and create a copy to the output
	err = UnblockFile(blockFile.ID, changedOutputFile)

	// No error
	c.Assert(err == nil, IsTrue)

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(changedOutputFile)

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == changedInputFileInfo.Size(), IsTrue)
}
