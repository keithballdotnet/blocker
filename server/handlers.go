package server

import (
	"encoding/json"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	"github.com/Inflatablewoman/blocker/crypto"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func GetHello(u *url.URL, h http.Header, _ interface{}) (int, http.Header, string, error) {
	log.Println("Got GET hello request")

	// Really simple hello
	return http.StatusOK, nil, "Server: Blocker", nil
}

// AuthorizeRequest - Will check the request authorization
func AuthorizeRequest(method string, u *url.URL, h http.Header) bool {

	date := h.Get("x-blocker-date")
	resource := u.Path
	authRequestKey := fmt.Sprintf("%s\n%s\n%s", method, date, resource)

	authorization := h.Get("Authorization")

	hmac := crypto.GetHmac256(authRequestKey, SharedKey)

	if authorization != hmac {
		log.Printf("Authorization FAILED: Auth: %s HMAC: %s RequestKey: \n%s", authorization, hmac, authRequestKey)
	}

	// Was the passed value the same as we expected?
	return authorization == hmac
}

// CopyHandler - The REST endpoint for deleting a BlockedFile
func CopyHandler(u *url.URL, h http.Header, _ interface{}) (int, http.Header, *blocks.BlockedFile, error) {
	log.Println("Got COPY block request")

	// Authoritze the request
	if !AuthorizeRequest("COPY", u, h) {
		return http.StatusUnauthorized, nil, nil, nil
	}

	itemID := u.Query().Get("itemID")

	blockedFile, err := blocks.CopyBlockedFile(itemID)
	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}

	// All good!
	return http.StatusOK, nil, &blockedFile, nil
}

// DeleteHandler - The REST endpoint for deleting a BlockedFile
func DeleteHandler(u *url.URL, h http.Header, _ interface{}) (int, http.Header, interface{}, error) {
	log.Println("Got DELETE block request")

	// Authoritze the request
	if !AuthorizeRequest("DELETE", u, h) {
		return http.StatusUnauthorized, nil, nil, nil
	}

	itemID := u.Query().Get("itemID")

	err := blocks.DeleteBlockedFile(itemID)
	if err != nil {
		return http.StatusInternalServerError, nil, false, err
	}

	// All good!
	return http.StatusNoContent, nil, nil, nil
}

// RawUploadHandler handles PUT operations
type RawUploadHandler struct {
}

func NewRawUploadHandler() RawUploadHandler {
	return RawUploadHandler{}
}

func (handler RawUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Got PUT upload request")

	// Authoritze the request
	if !AuthorizeRequest("PUT", r.URL, r.Header) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	BlockAndRespond(w, r.Body)
}

// PostMultipartUploadHandler handles POST operations
type PostMultipartUploadHandler struct{}

func NewPostMultipartUploadHandler() PostMultipartUploadHandler {
	return PostMultipartUploadHandler{}
}

func (handler PostMultipartUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Got POST upload request")

	// Authoritze the request
	if !AuthorizeRequest("POST", r.URL, r.Header) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(100000)
	if err != nil {
		HandleErrorWithResponse(w, err)
		return
	}

	m := r.MultipartForm

	// Currently we only support one file upload...  it will stop after processing the first file.
	for fname, _ := range m.File {
		files := m.File[fname]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, _ := files[i].Open()
			defer file.Close()
			if err != nil {
				HandleErrorWithResponse(w, err)
				return
			}

			fileName := files[i].Filename
			contentType := files[i].Header["Content-Type"][0]

			fmt.Printf("file: %#v type: %v\n", fileName, contentType)

			// This is stupid... but there you go.
			// See this for further discussion: http://www.reddit.com/r/golang/comments/2cdu7s/how_do_i_avoid_using_ioutilreadall/
			// fileBytes, err := ioutil.ReadAll(file)
			BlockAndRespond(w, file)
		}
	}
}

// Handle the uploaded data.
func BlockAndRespond(w http.ResponseWriter, content io.Reader) {

	// Create temp file
	outFile, err := ioutil.TempFile(os.TempDir(), "upload_")
	if err != nil {
		log.Println("Error serializing to json: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Save content to file
	io.Copy(outFile, content)

	// Close the file so it can be read
	outFile.Close()

	// Get some info about the file we are going test
	outputFileInfo, err := os.Stat(outFile.Name())
	if err != nil {
		log.Println("Error getting file info: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("File upload saved to \"%s\" was %v bytes", outputFileInfo.Name(), outputFileInfo.Size())

	// open the file and read the contents
	sourceFile, err := os.Open(outFile.Name())
	if err != nil {
		log.Println("Error serializing to json: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer sourceFile.Close()
	defer os.Remove(outFile.Name())

	blockedFile, err := blocks.BlockBuffer(sourceFile)

	w.WriteHeader(http.StatusCreated)
	w.Header()["Content-Type"] = []string{"application/json"}
	body, err := json.Marshal(blockedFile)
	if err != nil {
		log.Println("Error serializing to josn: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

type FileDownloadHandler struct {
}

func NewFileDownloadHandler() FileDownloadHandler {
	return FileDownloadHandler{}
}

func (handler FileDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Got GET file request")

	// Authoritze the request
	if !AuthorizeRequest("GET", r.URL, r.Header) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	itemID := r.URL.Query().Get("itemID")
	// fmt.Fprintf(w, "Going to get \"%v\"\n", itemID)

	buffer, err := blocks.UnblockFileToBuffer(itemID)

	if err != nil {
		HandleErrorWithResponse(w, err)
		return
	}

	header := w.Header()
	header["Content-Type"] = []string{"application/octet-stream"}
	// header["Content-Disposition"] = []string{"attachment;filename=" + fileName}

	buffer.WriteTo(w)

	//response, err := ioutil.ReadFile(outFile)
	//w.Write(response)
}

// checkClose is used to check the return from Close in a defer
// statement.
func checkClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}

func HandleErrorWithResponse(w http.ResponseWriter, error error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, error)
	return
}
