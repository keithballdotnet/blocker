package main

import (
	"fmt"
	. "github.com/Inflatablewoman/blocker/gocheck2"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type BlockerSuite struct{}

var _ = Suite(&BlockerSuite{})

func (s *BlockerSuite) TestGetFile(c *C) {
	fileId := "011a2bd1-cd46-4b00-a917-8e7268f805f3"

}
