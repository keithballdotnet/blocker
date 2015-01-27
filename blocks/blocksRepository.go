package blocks

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Inflatablewoman/azure"
	"github.com/couchbase/gomemcached/client"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// BlockRepository is the interface for saving blocks to disk
type BlockRepository interface {
	SaveBlock(bytes []byte, hash string) error
	GetBlock(blockHash string) ([]byte, error)
	CheckBlockExists(blockHash string) (bool, error)
	DeleteBlock(blockHash string) error
}

/* S3 Block Provider */

type S3BlockRepository struct {
	s3Store *s3.S3
	bucket  *s3.Bucket
}

// NewAzureBlockRepository
func NewS3BlockRepository() (S3BlockRepository, error) {

	key := os.Getenv("BLOCKER_S3_KEY")
	secret := os.Getenv("BLOCKER_S3_SECRET")
	bucketName := os.Getenv("BLOCKER_S3_BUCKET")

	if key == "" || secret == "" || bucketName == "" {
		panic("Enivronmental Variable: BLOCKER_S3_KEY or BLOCKER_S3_SECRET or BLOCKER_S3_BUCKET are empty!  You must set these values when using S3 storage!")
	}

	auth := aws.Auth{key, secret, ""}

	s3Store := s3.New(auth, aws.EUWest)
	// Create bucket...
	bucket := s3Store.Bucket(bucketName)

	// Presume bucket is created outside of this application
	/*err := bucket.PutBucket(s3.Private)
	if err != nil {
		log.Printf("Error creating bucket: %v", err)
	}*/

	s3BlockRepo := S3BlockRepository{s3Store, bucket}

	return s3BlockRepo, nil
}

func (r S3BlockRepository) SaveBlock(data []byte, blockHash string) error {

	// TODO:  Work out if there was an error here...
	err := r.bucket.Put(blockHash+".blk", data, "application/octet-stream", s3.Private)
	if err != nil {
		log.Printf("Error upload data: %v", err)
	}

	// log.Printf("Upload hash: %s Code: %v Size: %v", blockHash+".blk", res, len(data))

	return err
}

// Get a block from the repository
func (r S3BlockRepository) GetBlock(blockHash string) ([]byte, error) {

	// log.Printf("Download hash: %s Code: %v Size: %v", blockHash, res.StatusCode, len(contents))

	return r.bucket.Get(blockHash + ".blk")
}

// DeleteBlock - Deletes a block of data
func (r S3BlockRepository) DeleteBlock(blockHash string) error {

	return r.bucket.Del(blockHash + ".blk")
}

// Check to see if a block exists
func (r S3BlockRepository) CheckBlockExists(blockHash string) (bool, error) {

	res, err := r.bucket.Head(blockHash + ".blk")

	if err != nil {
		log.Printf("Get blob props err: %s blobs: %v", err, res)
		return false, err
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	// Block not present
	return false, nil
}

/* Azure BLOCK Provider */

type AzureBlockRepository struct {
	blobStore     azure.Azure
	containerName string
}

// NewAzureBlockRepository
func NewAzureBlockRepository() (AzureBlockRepository, error) {

	accountName := os.Getenv("BLOCKER_AZURE_ACCOUNT")
	secret := os.Getenv("BLOCKER_AZURE_SECRET")

	if accountName == "" || secret == "" {
		panic("Enivronmental Variable: BLOCKER_AZURE_ACCOUNT or BLOCKER_AZURE_SECRET are empty!  You must set these values when using azure storage!")
	}

	blobStore := azure.New(accountName, secret)

	azureBlockRepo := AzureBlockRepository{blobStore, "blocks"}

	return azureBlockRepo, nil
}

// Save persists a block into the repository
func (r AzureBlockRepository) SaveBlock(data []byte, blockHash string) error {
	buffer := bytes.NewBuffer(data)

	// TODO:  Work out if there was an error here...
	_, err := r.blobStore.FileUpload(r.containerName, blockHash+".blk", buffer)

	// log.Printf("Upload hash: %s Code: %v Size: %v", blockHash+".blk", res, len(data))

	return err
}

// Get a block from the repository
func (r AzureBlockRepository) GetBlock(blockHash string) ([]byte, error) {
	// Get data...
	res, err := r.blobStore.FileDownload(r.containerName, blockHash+".blk")
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// log.Printf("Download hash: %s Code: %v Size: %v", blockHash, res.StatusCode, len(contents))

	return contents, nil
}

// DeleteBlock - Deletes a block of data
func (r AzureBlockRepository) DeleteBlock(blockHash string) error {
	// Get data...
	_, err := r.blobStore.DeleteBlob(r.containerName, blockHash+".blk")
	if err != nil {
		return err
	}

	return nil
}

// Check to see if a block exists
func (r AzureBlockRepository) CheckBlockExists(blockHash string) (bool, error) {

	res, err := r.blobStore.GetProperties(r.containerName, blockHash+".blk")

	if err != nil {
		log.Printf("Get blob props err: %s blobs: %v", err, res)
		return false, err
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	// Block not present
	return false, nil
}

/* CouchBase BLOCK Provider */

type CouchBaseBlockRepository struct {
	bucket *couchbase.Bucket
}

// NewCouchBaseBlockRepository
func NewCouchBaseBlockRepository() (CouchBaseBlockRepository, error) {

	couchbaseEnvAddress := os.Getenv("CB_HOST")

	couchbaseAddress := "http://localhost:8091"
	if couchbaseEnvAddress != "" {
		couchbaseAddress = couchbaseEnvAddress
	}

	bucket, err := couchbase.GetBucket(couchbaseAddress, "default", "blocker")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		panic("Critical Error: No storage for files avilable in couchbase!")
	}

	log.Printf("NewCouchBaseBlockRepository: Connected to Couchbase Server: %s\n", couchbaseAddress)

	return CouchBaseBlockRepository{bucket}, nil
}

// Save persists a block into the repository
func (r CouchBaseBlockRepository) SaveBlock(bytes []byte, blockHash string) error {
	return r.bucket.SetRaw(blockHash, 0, bytes)
}

// Get a block from the repository
func (r CouchBaseBlockRepository) GetBlock(blockHash string) ([]byte, error) {

	if blockHash == "" {
		return nil, errors.New("No Block hash passed")
	}

	// Get data...
	blockData, err := r.bucket.GetRaw(blockHash)

	if err != nil {
		return nil, err
	}

	return blockData, nil
}

// DeleteBlock - Deletes a block of data
func (r CouchBaseBlockRepository) DeleteBlock(blockHash string) error {
	// Delete block
	return r.bucket.Delete(blockHash)
}

// Check to see if a block exists
func (r CouchBaseBlockRepository) CheckBlockExists(blockHash string) (bool, error) {

	// Check to see if hash is present
	result, err := r.bucket.Observe(blockHash)
	if err != nil {
		return false, err
	}

	// If the status is anything other than not found, then it's stored in couch base...
	if result.Status != memcached.ObservedNotFound {
		return true, nil
	}

	// Couch base does not have the block
	return false, nil
}

/* DISK BLOCK Provider */

// DiskBlockRepository : Saves blocks to disk
type DiskBlockRepository struct {
	path      string
	extension string
}

// NewBlockRepository
func NewDiskBlockRepository() (DiskBlockRepository, error) {

	// Use the path passed from ENV
	depositoryDir := os.Getenv("BLOCKER_DISK_DIR")
	if depositoryDir == "" {
		depositoryDir = filepath.Join(os.TempDir(), "blocker")
	}

	err := os.Mkdir(depositoryDir, 0777)
	if err != nil && !os.IsExist(err) {
		panic("Unable to create directory: " + err.Error())
	}

	log.Println("Storing blocks to: ", depositoryDir)

	return DiskBlockRepository{depositoryDir, ".blk"}, nil
}

func (r DiskBlockRepository) GetDataDirectory(hash string) (string, error) {

	dataDirectory := filepath.Join(r.path, string(hash[0]), string(hash[1]))

	// Does the directory aleady exist?
	exists, _ := directoryExists(dataDirectory)
	if exists {
		return dataDirectory, nil
	}

	dataDirectory = filepath.Join(r.path, string(hash[0]))
	err := os.Mkdir(dataDirectory, 0777)
	if err != nil && !os.IsExist(err) {
		return "", errors.New("Unable to create directory: " + err.Error())
	}

	dataDirectory = filepath.Join(r.path, string(hash[0]), string(hash[1]))
	err = os.Mkdir(dataDirectory, 0777)
	if err != nil && !os.IsExist(err) {
		return "", errors.New("Unable to create directory: " + err.Error())
	}

	return dataDirectory, nil
}

func directoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Save persists a block into the repository
func (r DiskBlockRepository) SaveBlock(bytes []byte, blockHash string) error {

	dataDirectory, _ := r.GetDataDirectory(blockHash)

	err := ioutil.WriteFile(filepath.Join(dataDirectory, blockHash+r.extension), bytes, 0644)
	if err != nil {
		log.Println(fmt.Sprintf("Error writing file : %v", err))
		return err
	}

	return nil
}

// Get a block from the repository
func (r DiskBlockRepository) GetBlock(blockHash string) ([]byte, error) {

	dataDirectory, _ := r.GetDataDirectory(blockHash)

	readBytes, err := ioutil.ReadFile(filepath.Join(dataDirectory, blockHash+r.extension))
	if err != nil {
		log.Println(fmt.Sprintf("Error reading block : %v", err))
		return nil, err
	}

	return readBytes, nil
}

// DeleteBlock - Deletes a block of data
func (r DiskBlockRepository) DeleteBlock(blockHash string) error {
	dataDirectory, _ := r.GetDataDirectory(blockHash)

	// Delete the file from disk...
	return os.Remove(filepath.Join(dataDirectory, blockHash+r.extension))
}

// Check to see if a block exists
func (r DiskBlockRepository) CheckBlockExists(blockHash string) (bool, error) {
	dataDirectory, _ := r.GetDataDirectory(blockHash)

	_, err := os.Stat(filepath.Join(dataDirectory, blockHash+r.extension))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/* FILEBlOCKINFO REPO */

// FileBlockInfoRepository inteface for FileBlockInfo storage
type BlockInfoRepository interface {
	SaveBlockInfo(blockInfo BlockInfo) error
	GetBlockInfo(hash string) (*BlockInfo, error)
	DeleteBlockInfo(hash string) error
}

var cbBlockInfoPrefix = "blocker:bi:"

// CouchbaseFileBlockInfoRepository is the couch base implementation of the FileBlockInfoRepository
type CouchbaseBlockInfoRepository struct {
	bucket         *couchbase.Bucket
	InMemoryBucket map[string]*BlockInfo
}

// NewBlockedFileRepository
func NewCouchbaseBlockInfoRepository() (CouchbaseBlockInfoRepository, error) {
	couchbaseEnvAddress := os.Getenv("CB_HOST")

	couchbaseAddress := "http://localhost:8091"
	if couchbaseEnvAddress != "" {
		couchbaseAddress = couchbaseEnvAddress
	}

	bucket, err := couchbase.GetBucket(couchbaseAddress, "default", "blocker")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		// NOTE:  I want this to run without a couchbase installation, so in event of error use a in memory store
		return CouchbaseBlockInfoRepository{nil, make(map[string]*BlockInfo)}, nil
	}

	log.Printf("NewCouchbaseFileBlockInfoRepository: Connected to Couchbase Server: %s\n", couchbaseAddress)

	return CouchbaseBlockInfoRepository{bucket, nil}, nil
}

func (r CouchbaseBlockInfoRepository) DeleteBlockInfo(hash string) error {
	if hash == "" {
		return errors.New("No Block File ID passed")
	}

	if r.bucket == nil {
		if _, ok := r.InMemoryBucket[hash]; ok {
			delete(r.InMemoryBucket, hash)
			return nil
		}

		return errors.New("Not found!")
	}

	if err := r.bucket.Delete(cbBlockInfoPrefix + hash); err != nil {
		return err
	}

	return nil
}

// Save persists a BlockedFile into the repository
func (r CouchbaseBlockInfoRepository) SaveBlockInfo(blockInfo BlockInfo) error {
	if r.bucket == nil {
		r.InMemoryBucket[blockInfo.Hash] = &blockInfo
		return nil
	}

	return r.bucket.Set(cbBlockInfoPrefix+blockInfo.Hash, 0, blockInfo)
}

// Get a BlockedFile from the repository
func (r CouchbaseBlockInfoRepository) GetBlockInfo(hash string) (*BlockInfo, error) {
	if hash == "" {
		return nil, errors.New("No hash passed")
	}

	if r.bucket == nil {
		if val, ok := r.InMemoryBucket[hash]; ok {
			return val, nil
		}

		return &BlockInfo{}, errors.New("Not found!")
	}

	var blockInfo BlockInfo

	if err := r.bucket.Get(cbBlockInfoPrefix+hash, &blockInfo); err != nil {
		return nil, err
	}

	return &blockInfo, nil
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

	bucket, err := couchbase.GetBucket(couchbaseAddress, "default", "blocker")
	if err != nil {
		log.Println(fmt.Sprintf("Error getting bucket:  %v", err))
		// NOTE:  I want this to run without a couchbase installation, so in event of error use a in memory store
		return BlockedFileRepository{nil, make(map[string]*BlockedFile)}, nil
	}

	log.Printf("NewBlockedFileRepository: Connected to Couchbase Server: %s\n", couchbaseAddress)

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
		if val, ok := r.InMemoryBucket[blockfileid]; ok {
			return val, nil
		}

		return &BlockedFile{}, errors.New("Not found!")
	}

	var blockedFile BlockedFile

	if err := r.bucket.Get(blockfileid, &blockedFile); err != nil {
		return nil, err
	}

	return &blockedFile, nil
}

// DeleteBlockedFile - Delete a blocked file
func (r BlockedFileRepository) DeleteBlockedFile(blockfileid string) error {
	if blockfileid == "" {
		return errors.New("No Block File ID passed")
	}

	if r.bucket == nil {
		if _, ok := r.InMemoryBucket[blockfileid]; ok {
			delete(r.InMemoryBucket, blockfileid)
			return nil
		}

		return errors.New("Not found!")
	}

	if err := r.bucket.Delete(blockfileid); err != nil {
		return err
	}

	return nil
}
