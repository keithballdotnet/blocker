package server

import (
	"encoding/json"
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	"io"
	"log"
	"net/http"
	"net/url"
)

func GetHello(u *url.URL, h http.Header, _ interface{}) (int, http.Header, string, error) {
	fmt.Println("Got GET hello request")

	// Really simple hello
	return http.StatusOK, nil, "hello", nil
}

type RawUploadHandler struct {
}

func NewRawUploadHandler() RawUploadHandler {
	return RawUploadHandler{}
}

func (handler RawUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header["Content-Type"][0]
	fileName := r.Header["Filename"][0]

	BlockAndRespond(w, fileName, contentType, r.Body)
}

type PostMultipartUploadHandler struct{}

func NewPostMultipartUploadHandler() PostMultipartUploadHandler {
	return PostMultipartUploadHandler{}
}

func (handler PostMultipartUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got POST upload request")

	file, header, err := r.FormFile("file") // the FormFile function takes in the POST input

	if err != nil {
		log.Println("Error reading input file: ", err)
		HandleErrorWithResponse(w, err)
		return
	}

	defer file.Close()

	fileName := header.Filename
	// TODO: Get content type
	contentType := "text/plain" //header.Header["Content-Type"][0]

	BlockAndRespond(w, fileName, contentType, file)
}

// Handle the uploaded data.
func BlockAndRespond(w http.ResponseWriter, filename string, contentType string, content io.Reader) {
	blockedFile, err := blocks.BlockBuffer(content, filename, contentType)

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
	fmt.Println("Got GET file request")

	itemID := r.URL.Query().Get("itemID")

	// fmt.Fprintf(w, "Going to get \"%v\"\n", itemID)

	buffer, err := blocks.UnblockFileToBuffer(itemID)

	if err != nil {
		HandleErrorWithResponse(w, err)
		return
	}

	header := w.Header()
	header["Content-Type"] = []string{"text/plain"}
	//header["Content-Disposition"] = []string{"attachment;filename=" + "document.txt"}

	buffer.WriteTo(w)

	//response, err := ioutil.ReadFile(outFile)
	//w.Write(response)
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
