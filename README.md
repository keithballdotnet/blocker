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

##Authorization

Authorization is done via a *Authorization* header sent in a request.  Anonymous requests are not allowed.  To authenticate a request, you must sign the request with the key for the account that is making the request and pass that signature as part of the request.  

Here you can see an example of a Authorization header
```
Authorization=RvPtP0QB7iIun1ehwheD4YUo7+fYfw7/ywl+HsC5Ddk=
```

You construct the signature is built in the following format:

```
authRequestSig = method + "\n" +
                 Date + "\n" +
                 resource
```

This would result in the following signature to be signed:

```
COPY\nWed, 28 Jan 2015 10:42:13 UTC\n/api/v1/blocker/6f90d707-3b6a-4321-b32c-3c1d37915c1b
```

Note that you MUST past the same date value in the request.  Date should be supplied in UTC using RFC1123 format.

```
x-blocker-date=Wed, 28 Jan 2015 10:42:13 UTC
```

  The signature must be exactly in the same order and include the new line character.  

Now encode the signature using the [HMAC-SHA256](http://en.wikipedia.org/wiki/Hash-based_message_authentication_code) algorithm using the shared key.

This will result in a key like this:
```
RvPtP0QB7iIun1ehwheD4YUo7+fYfw7/ywl+HsC5Ddk="
```

Example go code to create the signature

```go
date := time.Now().UTC().Format(time.RFC1123) // UTC time
request.Header.Add("x-blocker-date", date)

authRequestKey := fmt.Sprintf("%s\n%s\n%s", method, date, resource)

// See package http://golang.org/pkg/crypto/hmac/ on how golang creates hmacs
hmac := crypto.GetHmac256(authRequestKey, SharedKey)  

request.Header.Add("Authorization", hmac)
```

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
