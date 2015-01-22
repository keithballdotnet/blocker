package server

import (
	"encoding/json"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	. "github.com/Inflatablewoman/blocker/gocheck2"
	. "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ServerSuite struct{}

var _ = Suite(&ServerSuite{})

var baseURL = "http://localhost:8010"

// const inputFile = "testdata/tempest.txt"

func (s *ServerSuite) TestGetHello(c *C) {
	response, err := http.Get(baseURL + "/api/blocker")
	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	_, err = ioutil.ReadAll(response.Body)
	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
}

func (s *ServerSuite) TestFileUploadAndDownload(c *C) {

	// c.Skip("Just for now.  Will skip this.")

	const inputFile = "testdata/kjv.txt"
	outputFile := filepath.Join(os.TempDir(), "kjv.txt")

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)

	// open the file and read the contents
	sourceFile, err := os.Open(inputFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	defer sourceFile.Close()

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/blocker", baseURL), sourceFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	filename := inputFileInfo.Name()
	contentType := "text/plain"
	length := inputFileInfo.Size()

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
	//	c.Assert(blockedFile.Name == filename, IsTrue)
	c.Assert(blockedFile.Length == length, IsTrue)

	// Now try to get the data we uploaded
	response, err = http.Get(fmt.Sprintf("%s/api/blocker/%s", baseURL, blockedFile.ID))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Clean up any old file
	os.Remove(outputFile)

	outFile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	io.Copy(outFile, response.Body)
	response.Body.Close()
	outFile.Close()

	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Get some info about the file we are going test
	outputFileInfo, _ := os.Stat(outputFile)

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == inputFileInfo.Size(), IsTrue)

	// Copy the file
	request, err = http.NewRequest("COPY", fmt.Sprintf("%s/api/blocker/%s", baseURL, blockedFile.ID), nil)
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	var copiedBlockFile blocks.BlockedFile
	err = json.Unmarshal(body, &copiedBlockFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	fmt.Println(copiedBlockFile)

	c.Assert(copiedBlockFile.ID != blockedFile.ID, IsTrue, Commentf("Failed with error: %v", err))

	// Now try to get the data we copied
	response, err = http.Get(fmt.Sprintf("%s/api/blocker/%s", baseURL, copiedBlockFile.ID))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Clean up any old file
	os.Remove(outputFile)

	outFile, err = os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	defer outFile.Close()
	defer os.Remove(outputFile)

	io.Copy(outFile, response.Body)
	response.Body.Close()

	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Get some info about the file we are going test
	outputFileInfo, _ = os.Stat(outputFile)

	// Check we wrote the full file size
	c.Assert(outputFileInfo.Size() == inputFileInfo.Size(), IsTrue)

	// Delete the original upload
	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/blocker/%s", baseURL, blockedFile.ID), nil)
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	// Delete the copied upload
	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/blocker/%s", baseURL, copiedBlockFile.ID), nil)
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

}

func (s *ServerSuite) TestSimpleUploadAndDownload(c *C) {

	// Upload simple text
	uploadContent := "hello world"
	contentReader := strings.NewReader(uploadContent)

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/blocker", baseURL), contentReader)
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
	//	c.Assert(blockedFile.Name == filename, IsTrue)
	c.Assert(blockedFile.Length == length, IsTrue)

	// Now try to get the data we uploaded
	response, err = http.Get(fmt.Sprintf("%s/api/blocker/%s", baseURL, blockedFile.ID))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	receivedContent := string(body)
	c.Assert(receivedContent == uploadContent, IsTrue, Commentf("Content was: %v", receivedContent))

	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/blocker/%s", baseURL, blockedFile.ID), nil)
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
}
