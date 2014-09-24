package blocks

import (
	"fmt"
	. "github.com/Inflatablewoman/blocks/gocheck2"
	. "gopkg.in/check.v1"
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
	err, blockFile := CreateFile("/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest.txt")

	c.Assert(err == nil, IsTrue)
	c.Assert(blockFile.ID != "", IsTrue)
	c.Assert(len(blockFile.Blocks) > 0, IsTrue)

	fmt.Println("Created new file: ", blockFile)
}
