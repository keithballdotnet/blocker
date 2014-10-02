Blocker - A block based filesystem microservice in go
=======

[![GoDoc](https://godoc.org/github.com/Inflatablewoman/blocker?status.svg)](https://godoc.org/github.com/Inflatablewoman/blocker)
[![Build Status](https://drone.io/github.com/Inflatablewoman/blocker/status.png)](https://drone.io/github.com/Inflatablewoman/blocker/latest)

##Features

- Files are stored in blocks
- Blocks are hashed
- Blocks are encrypted
- Blocks are compressed using Snappy Compression
- Files where blocks have not changed reference old blocks
- A REST interface for downloading blocked documents
- A REST interface for uploading blocked documents
- Uses couchbase for BlockedFiles repository
- Uses local disk for FileBlocks repository

##REST API 

The REST API interface can be used to perform operations against the Filesystem.  Default location is localhost:8002.

HTTP Method | URI path | Description
------------|----------|------------
GET         | /api/blocker  | Retreives a hello.
GET         | /api/blocker/{itemID}  | Retreives a BlockedFile based on the passed ID
POST        | /api/blocker   | Uploads a file and returns a newly created BlockedFiles
PUT        | /api/blocker   | Uploads a file and returns a newly created BlockedFiles

##Example code
[Example test scenario](https://github.com/Inflatablewoman/blocker/blob/master/server/server_test.go)

##TODO

- Add a way to delete files
- Move encryption from TOY format to be a bit more secure
- Permissions
- Authentication
- Stream content to disk rather than save to temp location


