package server

import (
	"fmt"
	"github.com/Inflatablewoman/blocker/blocks"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func GetHello(u *url.URL, h http.Header, _ interface{}) (int, http.Header, string, error) {
	// Really simple hello
	return http.StatusOK, nil, "hello", nil
}

type FileDownloadHandler struct {
}

func NewFileDownloadHandler() FileDownloadHandler {
	return FileDownloadHandler{}
}

func (handler FileDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("itemID")

	// fmt.Fprintf(w, "Going to get \"%v\"\n", itemID)

	outFile := filepath.Join(os.TempDir(), "out.temp")

	// Clean up any old file
	os.Remove(outFile)

	// Get the file and create a copy to the output
	err := blocks.UnblockFile(itemID, outFile)

	if err != nil {
		HandleErrorWithResponse(w, err)
		return
	}

	header := w.Header()
	header["Content-Type"] = []string{"text/plain"}
	//header["Content-Disposition"] = []string{"attachment;filename=" + "document.txt"}

	response, err := ioutil.ReadFile(outFile)
	w.Write(response)
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
