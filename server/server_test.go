package server

import (
	"encoding/json"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	. "github.com/Inflatablewoman/blocker/gocheck2"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ServerSuite struct{}

var _ = Suite(&ServerSuite{})

var baseURL = "http://localhost:8002"

// const inputFile = "testdata/tempest.txt"

func (s *ServerSuite) TestGetHello(c *C) {
	response, err := http.Get(baseURL + "/api/blocks")
	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	_, err = ioutil.ReadAll(response.Body)
	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
}

func (s *ServerSuite) TestSimpleUploadAndDownload(c *C) {

	// Upload simple text
	uploadContent := "hello world"
	contentReader := strings.NewReader(uploadContent)

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/blocks/upload", baseURL), contentReader)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	filename := "helloWorld.txt"
	contentType := "text/plain"
	length := int64(len(uploadContent))

	request.Header.Add("FileName", filename)
	request.Header.Add("Content-Type", contentType)
	client := http.Client{}

	response, err := client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	var blockedFile blocks.BlockedFile
	err = json.Unmarshal(body, &blockedFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	fmt.Println(blockedFile)

	c.Assert(blockedFile.ID != "", IsTrue)
	c.Assert(blockedFile.Name == filename, IsTrue)
	c.Assert(blockedFile.ContentType == contentType, IsTrue)
	c.Assert(blockedFile.Length == length, IsTrue)

	// Now try to get the data we uploaded
	response, err = http.Get(fmt.Sprintf("%s/api/blocks/download/%s", baseURL, blockedFile.ID))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	receivedContent := string(body)
	c.Assert(receivedContent == uploadContent, IsTrue, Commentf("Content was: %v", receivedContent))
}
