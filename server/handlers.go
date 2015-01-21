package server

import (
	"encoding/json"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
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
	return http.StatusOK, nil, "hello", nil
}

// DeleteHandler - The REST endpoint for deleting a BlockedFile
func DeleteHandler(u *url.URL, h http.Header, _ interface{}) (int, http.Header, bool, error) {
	log.Println("Got DELETE block request")

	itemID := u.Query().Get("itemID")

	err := blocks.DeleteBlockedFile(itemID)
	if err != nil {
		return http.StatusInternalServerError, nil, false, err
	}

	// All good!
	return http.StatusOK, nil, true, nil
}

// RawUploadHandler handles PUT operations
type RawUploadHandler struct {
}

func NewRawUploadHandler() RawUploadHandler {
	return RawUploadHandler{}
}

func (handler RawUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Got PUT upload request")

	// contentType := r.Header["Content-Type"][0]
	// fileName := r.Header["Filename"][0]

	BlockAndRespond(w, r.Body)
}

// PostMultipartUploadHandler handles POST operations
type PostMultipartUploadHandler struct{}

func NewPostMultipartUploadHandler() PostMultipartUploadHandler {
	return PostMultipartUploadHandler{}
}

func (handler PostMultipartUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Got POST upload request")

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
	/*if util.IsNotFoundError(error) {
		w.WriteHeader(http.StatusNotFound)
		return
	}*/

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, error)
	return
}
