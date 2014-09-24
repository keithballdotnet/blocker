package blocks

import (
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"log"
)

// Repository : a Couchbase Server repository
type BlockRepository struct {
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

// Save persists a block into the repository
func (r BlockRepository) SaveBlock(block Block) error {
	return r.bucket.Set(block.ID, 0, block)
}

/*// Get retrieves an event sourced object by ID
func (r repository) Get(id string) (blocks.Block, error) {
	return nil, nil
}*/
