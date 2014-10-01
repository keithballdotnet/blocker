Block Filesystem in go
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

##TODO

- Add a way to delete files
- Document REST API
- Move encryption from TOY format to be a bit more secure


