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

##TODO

- Add REST inferface to file system
- Add a way to delete files


