package blocks

import (
	"fmt"
	. "github.com/Inflatablewoman/blocker/gocheck2"
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

const inputFile = "testdata/tempest.txt"
const bibleInFile = "testdata/kjv.txt"
const bibleOutFileName = "kjv.txt"
const changedInputFile = "testdata/tempest_changed.txt"
const changedAgainInputFile = "/testdata/tempest_changed_again.txt"
const outputFileName = "tempest.txt"
const changedOutputFileName = "tempest_changed.txt"
const liteIdeInFile = "testdata/liteidex23.2.linux-32.tar.bz2"
const liteIdeoutFile = "liteidex23.2.linux-32.tar.bz2"

func (s *BlockSuite) TestLiteide(c *C) {

	c.Skip("Want to work out what is going on")

	/*BlockSize = BlockSize4Mb
	UseCompression = true
	UseEncryption = true*/

	// Block the bigger
	start := time.Now()
	bibleBlockFile, err := BlockFile(liteIdeInFile)
	end := time.Now()

	fmt.Printf("Blocked LiteIde took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	bibleOutFile := os.TempDir() + "/" + liteIdeoutFile

	// Clean up any old file
	os.Remove(bibleOutFile)

	// Get the file and create a copy to the output
	start = time.Now()
	err = UnblockFile(bibleBlockFile.ID, bibleOutFile)
	end = time.Now()

	fmt.Printf("Unblocked LiteIde took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
}

func (s *BlockSuite) TestKingJamesBible(c *C) {

	c.Skip("Want to work out what is going on")

	/*BlockSize = BlockSize4Mb
	UseCompression = true
	UseEncryption = true*/

	// Block the bigger
	start := time.Now()
	bibleBlockFile, err := BlockFile(bibleInFile)
	end := time.Now()

	fmt.Printf("Blocked King James Bible took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	bibleOutFile := os.TempDir() + "/" + bibleOutFileName

	// Clean up any old file
	os.Remove(bibleOutFile)

	// Get the file and create a copy to the output
	start = time.Now()
	err = UnblockFile(bibleBlockFile.ID, bibleOutFile)
	end = time.Now()

	fmt.Printf("Unblocked King James Bible took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

}

func (s *BlockSuite) TestTempest(c *C) {

	/*BlockSize = BlockSize4Mb
	UseCompression = true
	UseEncryption = true*/

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)

	// Block the file
	start := time.Now()
	blockFile, err := BlockFile(inputFile)
	end := time.Now()

	fmt.Printf("Blocked Tempest took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// Check it was created in the past
	c.Assert(blockFile.Created.Before(time.Now()), IsTrue)

	// Check we read the full file size
	c.Assert(blockFile.Length == inputFileInfo.Size(), IsTrue)

	// Make sure the item returned some blocks
	c.Assert(len(blockFile.BlockList) > 0, IsTrue)

	outputFile := os.TempDir() + "/" + outputFileName

	// Clean up any old file
	os.Remove(outputFile)

	// Get the file and create a copy to the output
	start = time.Now()
	err = UnblockFile(blockFile.ID, outputFile)
	end = time.Now()

	fmt.Printf("Unblocked Tempest took: %v\n", end.Sub(start))

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(outputFile)

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == inputFileInfo.Size(), IsTrue)
}

// Benchmark the blocking of Tempest file
func (s *BlockSuite) BenchmarkTempestCompressedEncrypted30Kb(c *C) {
	for i := 0; i < c.N; i++ {
		// Set up test
		BlockSize = BlockSize30Kb
		UseCompression = true
		UseEncryption = true
		BlockFile(inputFile)
	}
}

func (s *BlockSuite) BenchmarkTempestCompressedEncrypted4Mb(c *C) {
	for i := 0; i < c.N; i++ {
		// Set up test
		BlockSize = BlockSize4Mb
		UseCompression = true
		UseEncryption = true
		BlockFile(inputFile)
	}
}

func (s *BlockSuite) BenchmarkTempestUncompressedUnencrypted4Mb(c *C) {
	for i := 0; i < c.N; i++ {
		// Set up test
		BlockSize = BlockSize4Mb
		UseCompression = false
		UseEncryption = false
		// Need to clean out the block store directory

		BlockFile(inputFile)
	}
}

func (s *BlockSuite) TestChangeTempest(c *C) {

	c.Skip("Not what I want to test right now")

	// Set up test
	BlockSize = BlockSize30Kb
	UseCompression = true
	UseEncryption = true

	// Get some info about the file we are going test
	changedInputFileInfo, _ := os.Stat(changedInputFile)

	blockFile, err := BlockFile(inputFile)

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	firstFileHash := blockFile.BlockList[0].Hash

	// Block the file again.
	blockFile, err = BlockFile(inputFile)

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	// Check that block used in first block is the same
	c.Assert(firstFileHash == blockFile.BlockList[0].Hash, IsTrue)

	// Block the file again.  New version should be created
	blockFile, err = BlockFile(changedInputFile)

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Check we have an ID
	c.Assert(blockFile.ID != "", IsTrue)

	changedOutputFile := os.TempDir() + "/" + changedOutputFileName

	// Clean up any old file
	os.Remove(changedOutputFile)

	// Get the file and create a copy to the output
	err = UnblockFile(blockFile.ID, changedOutputFile)

	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(changedOutputFile)

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == changedInputFileInfo.Size(), IsTrue)
}

/*func (s *BlockSuite) TestBlockRepository(c *C) {

	// Get repository
	repository, err := NewBlockRepository()
	c.Assert(err == nil, IsTrue)

	hash := "sha256:justalonghashthatgoesonandonandon"
	data := []byte(hash)

	// Create a block
	block := Block{hash, data}

	// Save
	err = repository.SaveBlock(block)
	c.Assert(err == nil, IsTrue)

}*/
