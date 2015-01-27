Blocker 
=======
A block based filesystem microservice written in go

[![GoDoc](https://godoc.org/github.com/Inflatablewoman/blocker?status.svg)](https://godoc.org/github.com/Inflatablewoman/blocker)
[![Build Status](https://travis-ci.org/Inflatablewoman/blocker.svg)](https://travis-ci.org/Inflatablewoman/blocker.svg)
[![Coverage Status](https://coveralls.io/repos/Inflatablewoman/blocker/badge.svg)](https://coveralls.io/r/Inflatablewoman/blocker)

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
- Possible to specify a backend Storage provider
   + nfs - Local mount disk storage (GlusterFS could be used)
   + couchbase - Couchbase Raw Binary storage
   + azure - Azure Simple Storage
   + s3 - s3 storage

##REST API 

The REST API interface can be used to perform operations against the Filesystem.  Default location is localhost:8010.

HTTP Method | URI path | Description
------------|----------|------------
GET         | /api/blocker  | Retrieves a hello.
GET         | /api/blocker/{itemID}  | Retrieves a BlockedFile based on the passed ID
DELETE      | /api/blocker/{itemID}  | Delete a BlockedFile based on the passed ID
COPY      | /api/blocker/{itemID}  | Creates a copy of a BlockedFile based on the passed ID
POST        | /api/blocker   | Uploads a file and returns a newly created BlockedFile
PUT        | /api/blocker   | Uploads a file and returns a newly created BlockedFile

##Example code
[Example test scenario](https://github.com/Inflatablewoman/blocker/blob/master/server/server_test.go)

##TODO

- Store Symmetric keys in a different location from that of the master key
- The store should not use the hash as the ID as this would then be a predictable location for a file.  A problem?
- Permissions
- Authentication
