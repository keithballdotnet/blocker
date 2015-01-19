package blocks

import (
	"errors"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// BlockRepository is the interface for saving blocks to disk
type BlockRepository interface {
	SaveBlock(bytes []byte, hash string) error
	GetBlock(blockHash string) ([]byte, error)
	CheckBlockExists(blockHash string) (bool, error)
}

/* DISK BLOCK Provider */

// DiskBlockRepository : Saves blocks to disk
type DiskBlockRepository struct {
	path      string
	extension string
}

// NewBlockRepository
func NewDiskBlockRepository() (DiskBlockRepository, error) {

	depositoryDir := filepath.Join(os.TempDir(), "blocker")

	err := os.Mkdir(depositoryDir, 0777)
	if err != nil && !os.IsExist(err) {
		panic("Unable to create directory: " + err.Error())
	}

	log.Println("Storing blocks to: ", depositoryDir)

	return DiskBlockRepository{depositoryDir, ".blk"}, nil
}

// Save persists a block into the repository
func (r DiskBlockRepository) SaveBlock(bytes []byte, hash string) error {
	/*bytes, err := json.Marshal(block)
	if err != nil {
		log.Println(fmt.Sprintf("Error marshalling file : %v", err))
		return err
	}*/

	err := ioutil.WriteFile(filepath.Join(r.path, hash+r.extension), bytes, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("Error writing file : %v", err))
		return err
	}

	return nil
}

// Get a block from the repository
func (r DiskBlockRepository) GetBlock(blockHash string) ([]byte, error) {

	readBytes, err := ioutil.ReadFile(filepath.Join(r.path, blockHash+r.extension))
	if err != nil {
		log.Println(fmt.Sprintf("Error reading block : %v", err))
		return nil, err
	}

	// json.Unmarshal(readBytes, &block)

	return readBytes, nil
}

// Check to see if a block exists
func (r DiskBlockRepository) CheckBlockExists(blockHash string) (bool, error) {
	_, err := os.Stat(filepath.Join(r.path, blockHash+r.extension))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// BlockedFileRepository : a Couchbase Server repository
type BlockedFileRepository struct {
	bucket         *couchbase.Bucket
	InMemoryBucket map[string]*BlockedFile
}

// NewBlockedFileRepository
func NewBlockedFileRepository() (BlockedFileRepository, error) {
	couchbaseEnvAddress := os.Getenv("CB_HOST")

	couchbaseAddress := "http://localhost:8091"
	if couchbaseEnvAddress != "" {
		couchbaseAddress = couchbaseEnvAddress
	}

	bucket, err := couchbase.GetBucket(couchbaseAddress, "default", "blockedfiles")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		// NOTE:  I want this to run without a couchbase installation, so in event of error use a in memory store
		return BlockedFileRepository{nil, make(map[string]*BlockedFile)}, nil
	}

	log.Printf("Connected to Couchbase Server: %s\n", couchbaseAddress)

	return BlockedFileRepository{bucket, nil}, nil
}

// Save persists a BlockedFile into the repository
func (r BlockedFileRepository) SaveBlockedFile(blockedFile BlockedFile) error {
	if r.bucket == nil {
		r.InMemoryBucket[blockedFile.ID] = &blockedFile
		return nil
	}

	return r.bucket.Set(blockedFile.ID, 0, blockedFile)
}

// Get a BlockedFile from the repository
func (r BlockedFileRepository) GetBlockedFile(blockfileid string) (*BlockedFile, error) {
	if blockfileid == "" {
		return nil, errors.New("No Block File ID passed")
	}

	if r.bucket == nil {
		return r.InMemoryBucket[blockfileid], nil
	}

	var blockedFile BlockedFile

	if err := r.bucket.Get(blockfileid, &blockedFile); err != nil {
		return nil, err
	}

	return &blockedFile, nil
}
