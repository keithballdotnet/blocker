package blocks

import (
	"errors"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"log"
)

// BlockRepository : a Couchbase Server repository
type BlockRepository struct {
	bucket *couchbase.Bucket
}

// BlockedFileRepository : a Couchbase Server repository
type BlockedFileRepository struct {
	bucket *couchbase.Bucket
}

// NewBlockRepository
func NewBlockRepository() (BlockRepository, error) {
	c, err := couchbase.Connect("http://localhost:8091/")
	if err != nil {
		log.Println(fmt.Sprintf("Error connecting to couchbase : %v", err))
		return BlockRepository{}, err
	}

	pool, err := c.GetPool("default")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting pool:  %v", err))
		return BlockRepository{}, err
	}

	bucket, err := pool.GetBucket("blocks")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		return BlockRepository{}, err
	}

	return BlockRepository{bucket}, nil
}

// NewBlockedFileRepository
func NewBlockedFileRepository() (BlockedFileRepository, error) {
	c, err := couchbase.Connect("http://localhost:8091/")
	if err != nil {
		log.Println(fmt.Sprintf("Error connecting to couchbase : %v", err))
		return BlockedFileRepository{}, err
	}

	pool, err := c.GetPool("default")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting pool:  %v", err))
		return BlockedFileRepository{}, err
	}

	bucket, err := pool.GetBucket("blockedfiles")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		return BlockedFileRepository{}, err
	}

	return BlockedFileRepository{bucket}, nil
}

// Save persists a block into the repository
func (r BlockRepository) SaveBlock(block Block) error {
	return r.bucket.Set(block.ID, 0, block)
}

// Get a block from the repository
func (r BlockRepository) GetBlock(blockid string) (*Block, error) {
	var block Block

	if err := r.bucket.Get(blockid, &block); err != nil {
		return nil, err
	}

	return &block, nil
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
