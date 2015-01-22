Blocker 
=======
A block based filesystem microservice written in go
[![GoDoc](https://godoc.org/github.com/Inflatablewoman/blocker?status.svg)](https://godoc.org/github.com/Inflatablewoman/blocker)
[![Build Status](https://travis-ci.org/Inflatablewoman/blocker.svg)](https://travis-ci.org/Inflatablewoman/blocker.svg)

##Features

- Files are stored in blocks
- Immutable blocks (Append Only)
- Blocks are hashed
- Blocks are encrypted with own unique Symmetric Key (Symmetric Keys are Encrypted with RSA Master Key)
- Blocks are compressed using Snappy Compression
- Reduces data duplication
- Files where blocks have not changed reference old blocks
- A REST interface for manipulating blocks
- Uses couchbase for BlockedFiles repository
- Possible to specify mulitple Storage Providers.
   + nfs - Local mount disk storage
   + couchbase - Couchbase Raw Binary storage
   + azure - Azure Simple Storage
   + s3 - Planned s3 storage

##REST API 

The REST API interface can be used to perform operations against the Filesystem.  Default location is localhost:8010.

HTTP Method | URI path | Description
------------|----------|------------
GET         | /api/blocker  | Retrieves a hello.
GET         | /api/blocker/{itemID}  | Retrieves a BlockedFile based on the passed ID
DELETE      | /api/blocker/{itemID}  | Delete a BlockedFile based on the passed ID
POST        | /api/blocker   | Uploads a file and returns a newly created BlockedFiles
PUT        | /api/blocker   | Uploads a file and returns a newly created BlockedFiles

##Example code
[Example test scenario](https://github.com/Inflatablewoman/blocker/blob/master/server/server_test.go)

##TODO

- Move encryption from TOY format to be a bit more secure
- Permissions
- Authentication