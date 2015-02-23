package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	"github.com/Inflatablewoman/blocker/crypto"
	. "github.com/Inflatablewoman/blocker/gocheck2"
	. "gopkg.in/check.v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test(t *testing.T) {
	TestingT(t)
}

type ServerSuite struct{}

var _ = Suite(&ServerSuite{})

var baseURL = "https://localhost:8010"

var testAuthKey = "e7yflbeeid26rredmwtbiyzxijzak6altcnrsi4yol2f5sexbgdwevlpgosfoeyy"

// const inputFile = "testdata/tempest.txt"
func (s *ServerSuite) TestSetupAuthenticationKey(c *C) {
	filepath.Join(os.TempDir(), "blocker")
	defaultAuthDir := filepath.Join(os.TempDir(), "blocker")
	err := os.Mkdir(defaultAuthDir, 0777)
	if err != nil && !os.IsExist(err) {
		c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	}

	keyPath := filepath.Join(defaultAuthDir, "blockertest.key")

	testAuthKey := "e7yflbeeid26rredmwtbiyzxijzak6altcnrsi4yol2f5sexbgdwevlpgosfoeyy"
	// Write key to key file
	err = ioutil.WriteFile(keyPath, []byte(testAuthKey), 0644)

	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Set the key path
	flag.Set("sharedKey", keyPath)

	// Load the key
	SetupAuthenticationKey()

	c.Assert(SharedKey == testAuthKey, IsTrue, Commentf("Wanted key: %v Got Key", SharedKey, testAuthKey))

	// Clean up the key file
	os.Remove(keyPath)
}

func (s *ServerSuite) TestGetHello(c *C) {
	response, err := http.Get(baseURL + "/api/v1/blocker")
	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	_, err = ioutil.ReadAll(response.Body)
	// No error
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
}

// SetAuth will set blocker auth headers
func SetAuth(request *http.Request, method string, resource string) *http.Request {

	date := time.Now().UTC().Format(time.RFC1123) // UTC time
	request.Header.Add("x-blocker-date", date)

	authRequestKey := fmt.Sprintf("%s\n%s\n%s", method, date, resource)

	hmac := crypto.GetHmac256(authRequestKey, SharedKey)

	//fmt.Printf("SharedKey: %s HMAC: %s RequestKey: \n%s\n", SharedKey, hmac, authRequestKey)

	request.Header.Add("Authorization", hmac)

	return request
}

func (s *ServerSuite) TestFileUploadAndDownload(c *C) {

	// Set the key path  Make sure the default key is loaded.
	flag.Set("sharedKey", "")

	// Load the key
	SetupAuthenticationKey()

	// c.Skip("Just for now.  Will skip this.")

	const inputFile = "testdata/kjv.txt"
	outputFile := filepath.Join(os.TempDir(), "kjv.txt")

	// Get some info about the file we are going test
	inputFileInfo, _ := os.Stat(inputFile)

	// open the file and read the contents
	sourceFile, err := os.Open(inputFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	defer sourceFile.Close()

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/blocker", baseURL), sourceFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Set auth
	request = SetAuth(request, "PUT", "/api/v1/blocker")

	filename := inputFileInfo.Name()
	contentType := "text/plain"
	length := inputFileInfo.Size()

	request.Header.Add("FileName", filename)
	request.Header.Add("Content-Type", contentType)
	client := http.Client{}

	response, err := client.Do(request)

	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusCreated, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	var blockedFile blocks.BlockedFile
	err = json.Unmarshal(body, &blockedFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	fmt.Println(FormatJSON("BlockedFile", blockedFile))

	c.Assert(blockedFile.ID != "", IsTrue)
	//	c.Assert(blockedFile.Name == filename, IsTrue)
	c.Assert(blockedFile.Length == length, IsTrue)

	// Now try to get the data we uploaded
	request, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)
	// Set auth
	request = SetAuth(request, "GET", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))
	// Now try to get the data we copied
	response, err = client.Do(request)
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

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
	request, err = http.NewRequest("COPY", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)

	// Set auth
	request = SetAuth(request, "COPY", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))

	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	var copiedBlockFile blocks.BlockedFile
	err = json.Unmarshal(body, &copiedBlockFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	fmt.Println(FormatJSON("BlockedFile", copiedBlockFile))

	c.Assert(copiedBlockFile.ID != blockedFile.ID, IsTrue, Commentf("Failed with error: %v", err))

	request, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, copiedBlockFile.ID), nil)
	// Set auth
	request = SetAuth(request, "GET", fmt.Sprintf("/api/v1/blocker/%s", copiedBlockFile.ID))

	// Now try to get the data we copied
	response, err = client.Do(request)
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
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
	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)
	// Set auth
	request = SetAuth(request, "DELETE", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusNoContent, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

	// Delete the copied upload
	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, copiedBlockFile.ID), nil)
	// Set auth
	request = SetAuth(request, "DELETE", fmt.Sprintf("/api/v1/blocker/%s", copiedBlockFile.ID))
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusNoContent, IsTrue, Commentf("Failed with status: %v", response.StatusCode))

}

// PrintJSON prints JSON with correct indents useful for log outputs.
func FormatJSON(name string, thing interface{}) string {
	jsonBytes, err := json.MarshalIndent(thing, "", "    ")
	if err != nil {
		return fmt.Sprintf("ERROR: %s\n", err)
	}
	return fmt.Sprintf("%s: %s\n", name, string(jsonBytes))
}

func (s *ServerSuite) TestAuthFail(c *C) {

	// Set the key path  Make sure the default key is loaded.
	flag.Set("sharedKey", "")

	// Load the key
	SetupAuthenticationKey()

	// Upload simple text
	uploadContent := "hello world"
	contentReader := strings.NewReader(uploadContent)

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/blocker", baseURL), contentReader)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Set incorrect method...
	request = SetAuth(request, "COPY", "/api/v1/blocker")
	filename := "helloWorld.txt"
	contentType := "text/plain"
	request.Header.Add("FileName", filename)
	request.Header.Add("Content-Type", contentType)
	client := http.Client{}

	response, err := client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusUnauthorized, IsTrue, Commentf("Expected AD got: %v", response.StatusCode))

	request, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/blocker", baseURL), contentReader)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	// Set incorrect resource
	request = SetAuth(request, "PUT", "/api/v1/blocker/block")
	request.Header.Add("FileName", filename)
	request.Header.Add("Content-Type", contentType)

	response, err = client.Do(request)
	c.Assert(response.StatusCode == http.StatusUnauthorized, IsTrue, Commentf("Expected AD got: %v", response.StatusCode))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
}

func (s *ServerSuite) TestSimpleUploadAndDownload(c *C) {

	// Set the key path  Make sure the default key is loaded.
	flag.Set("sharedKey", "")

	// Load the key
	SetupAuthenticationKey()

	// Upload simple text
	uploadContent := "hello world"
	contentReader := strings.NewReader(uploadContent)

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/blocker", baseURL), contentReader)
	request = SetAuth(request, "PUT", "/api/v1/blocker")

	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	filename := "helloWorld.txt"
	contentType := "text/plain"
	length := int64(len(uploadContent))

	request.Header.Add("FileName", filename)
	request.Header.Add("Content-Type", contentType)
	client := http.Client{}

	response, err := client.Do(request)
	c.Assert(response.StatusCode == http.StatusCreated, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	var blockedFile blocks.BlockedFile
	err = json.Unmarshal(body, &blockedFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	fmt.Println(FormatJSON("BlockedFile", blockedFile))

	c.Assert(blockedFile.ID != "", IsTrue)
	//	c.Assert(blockedFile.Name == filename, IsTrue)
	c.Assert(blockedFile.Length == length, IsTrue)

	// Now try to get the data we uploaded
	request, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)
	request = SetAuth(request, "GET", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))
	response, err = client.Do(request)
	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	receivedContent := string(body)
	c.Assert(receivedContent == uploadContent, IsTrue, Commentf("Content was: %v", receivedContent))

	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)
	request = SetAuth(request, "DELETE", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusNoContent, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
}

func (s *ServerSuite) TestLarge240MbUploadAndDownload(c *C) {

	c.Skip("Skip large upload test")

	// Set the key path  Make sure the default key is loaded.
	flag.Set("sharedKey", "")

	// Load the key
	SetupAuthenticationKey()

	// Upload simple text
	uploadContent := crypto.RandomSecret(150000000)
	contentReader := strings.NewReader(uploadContent)

	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/blocker", baseURL), contentReader)
	request = SetAuth(request, "PUT", "/api/v1/blocker")

	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	filename := "helloWorld.txt"
	contentType := "text/plain"
	length := int64(len(uploadContent))

	request.Header.Add("FileName", filename)
	request.Header.Add("Content-Type", contentType)
	client := http.Client{}

	start := time.Now()
	response, err := client.Do(request)
	end := time.Now()

	fmt.Printf("Got Response: Large file BLOCK took: %v\n", end.Sub(start))

	c.Assert(response.StatusCode == http.StatusCreated, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	var blockedFile blocks.BlockedFile
	err = json.Unmarshal(body, &blockedFile)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	fmt.Println(FormatJSON("BlockedFile", blockedFile))

	c.Assert(blockedFile.ID != "", IsTrue)
	//	c.Assert(blockedFile.Name == filename, IsTrue)
	c.Assert(blockedFile.Length == length, IsTrue)

	// Now try to get the data we uploaded
	request, err = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)
	request = SetAuth(request, "GET", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))
	start = time.Now()
	response, err = client.Do(request)
	end = time.Now()

	fmt.Printf("Got Response: Large file BLOCK took: %v\n", end.Sub(start))

	c.Assert(response.StatusCode == http.StatusOK, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))

	receivedContent := string(body)
	c.Assert(receivedContent == uploadContent, IsTrue, Commentf("Content was: %v", receivedContent))

	request, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/blocker/%s", baseURL, blockedFile.ID), nil)
	request = SetAuth(request, "DELETE", fmt.Sprintf("/api/v1/blocker/%s", blockedFile.ID))
	response, err = client.Do(request)
	c.Assert(err == nil, IsTrue, Commentf("Failed with error: %v", err))
	c.Assert(response.StatusCode == http.StatusNoContent, IsTrue, Commentf("Failed with status: %v", response.StatusCode))
}
