package blocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// BlockRepository : Saves blocks to disk
type BlockRepository struct {
	path string
}

// NewBlockRepository
func NewBlockRepository() (BlockRepository, error) {

	depositoryDir := filepath.Join(os.TempDir(), "blocks")

	err := os.Mkdir(depositoryDir, 0777)
	if err != nil && !os.IsExist(err) {
		panic("Unable to create directory: " + err.Error())
	}

	return BlockRepository{depositoryDir}, nil
}

// Save persists a block into the repository
func (r BlockRepository) SaveBlock(block Block) error {
	bytes, err := json.Marshal(block)
	if err != nil {
		log.Println(fmt.Sprintf("Error marshalling file : %v", err))
		return err
	}

	err = ioutil.WriteFile(filepath.Join(r.path, block.Hash+".json"), bytes, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("Error writing file : %v", err))
		return err
	}

	return nil
}

// Get a block from the repository
func (r BlockRepository) GetBlock(blockHash string) (*Block, error) {
	var block Block

	readBytes, err := ioutil.ReadFile(filepath.Join(r.path, blockHash+".json"))
	if err != nil {
		log.Println(fmt.Sprintf("Error reading block : %v", err))
		return nil, err
	}

	json.Unmarshal(readBytes, &block)

	return &block, nil
}

// Check to see if a block exists
func (r BlockRepository) CheckBlockExists(blockHash string) (bool, error) {
	_, err := os.Stat(r.path + blockHash + ".json")
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
	bucket, err := couchbase.GetBucket("http://localhost:8091", "default", "blockedfiles")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		// NOTE:  I want this to run without a couchbase installation, so in event of error use a in memory store
		return BlockedFileRepository{nil, make(map[string]*BlockedFile)}, nil
	}

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
