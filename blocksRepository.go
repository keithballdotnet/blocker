package blocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"io/ioutil"
	"log"
	"os"
)

// BlockRepository : Saves blocks to disk
type BlockRepository struct {
	path string
}

// NewBlockRepository
func NewBlockRepository() (BlockRepository, error) {

	depositoryDir := os.TempDir() + "/blocks/"

	os.MkdirAll(depositoryDir, os.ModeDir)

	return BlockRepository{depositoryDir}, nil
}

// Save persists a block into the repository
func (r BlockRepository) SaveBlock(block Block) error {
	bytes, err := json.Marshal(block)
	if err != nil {
		log.Println(fmt.Sprintf("Error marshalling file : %v", err))
		return err
	}

	err = ioutil.WriteFile(r.path+block.Hash+".json", bytes, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("Error writing file : %v", err))
		return err
	}

	return nil
}

// Get a block from the repository
func (r BlockRepository) GetBlock(blockHash string) (*Block, error) {
	var block Block

	readBytes, err := ioutil.ReadFile(r.path + blockHash + ".json")
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
	bucket *couchbase.Bucket
}

// NewBlockedFileRepository
func NewBlockedFileRepository() (BlockedFileRepository, error) {
	bucket, err := couchbase.GetBucket("http://localhost:8091", "default", "blockedfiles")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		return BlockedFileRepository{}, err
	}

	return BlockedFileRepository{bucket}, nil
}

// Save persists a BlockedFile into the repository
func (r BlockedFileRepository) SaveBlockedFile(blockedFile BlockedFile) error {
	return r.bucket.Set(blockedFile.ID, 0, blockedFile)
}

// Get a BlockedFile from the repository
func (r BlockedFileRepository) GetBlockedFile(blockfileid string) (*BlockedFile, error) {
	if blockfileid == "" {
		return nil, errors.New("No Block File ID passed")
	}

	var blockedFile BlockedFile

	if err := r.bucket.Get(blockfileid, &blockedFile); err != nil {
		return nil, err
	}

	return &blockedFile, nil
}
