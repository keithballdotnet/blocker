Blocker
=======
A block based filesystem microservice written in go

[![GoDoc](https://godoc.org/github.com/Inflatablewoman/blocker?status.svg)](https://godoc.org/github.com/Inflatablewoman/blocker)
[![Build Status](https://travis-ci.org/Inflatablewoman/blocker.svg)](https://travis-ci.org/Inflatablewoman/blocker)
[![Coverage Status](https://coveralls.io/repos/Inflatablewoman/blocker/badge.svg)](https://coveralls.io/r/Inflatablewoman/blocker)

##The case for Blocker

###What is blocker?

Blocker is a block based file storage microservice. Files are stored in blocks referenced by their hash value. Blocks are encrypted. Blocks are compressed. Blocker is storage provider agnostic. Blocker can save storage space by removing duplication of data.

Imagine the following upload of a 14MB video file. The next time this document is uploaded or copied in the system then blocker will check for the existence of the hashes h1, h2, h3 and h4. If the hashes exist then no action will be taken.

![](images/DropboxFileFormat.png?raw=true)

_Image taken from blogs.dropbox.com_

If a change is made to the video file, for example an extra 1mb of data is appended and uploaded. Then the document will now exist as h1, h2, h3 and the newly created h5. h4 will remain stored and be returned if the old version of the document is requested.

The basis of this is taken from a [2014 tech blog from Dropbox](https://blogs.dropbox.com/tech/2014/07/streaming-file-synchronization/).

##Features

- Files are stored in blocks
- Immutable blocks (Append Only Data Store)
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
   + s3 - Amazon s3 storage

##Todo

- Updating a block from the block list

##Authorization

Authorization is done via a *Authorization* header sent in a request.  Anonymous requests are not allowed.  To authenticate a request, you must sign the request with the shared key when making the request and pass that signature as part of the request.  

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

##Data Encryption

Data encryption is done using [openpgp golang library](https://godoc.org/golang.org/x/crypto/openpgp).  Specifically SHA256 hashes and the AES256 Cipher is used for encryption.  Compression is currently handled seperatley, using google's [Snappy compression](https://code.google.com/p/snappy/).

```go
// Default encryption settings (No encryption done by pgp)
var pgpConfig = &packet.Config{
	DefaultHash:            crypto.SHA256,
	DefaultCipher:          packet.CipherAES256,
	DefaultCompressionAlgo: packet.CompressionNone,
}
```

Data encryption requires a key, to generate a key you can use the following command.  

```
gpg2 --batch --gen-key --armor ./src/github.com/Inflatablewoman/blocker/crypto/gpg.batch
```

Once you have the public and private key files, you should set the following OS variables to the path of where the keys can be found.

```
export BLOCKER_PGP_PUBLICKEY=path/to/.pubring.gpg
export BLOCKER_PGP_PRIVATEKEY=path/to/.secring.gpg
```

A good explanation of PGP Encryption can be found on [wikipedia](http://en.wikipedia.org/wiki/Pretty_Good_Privacy).  In essence each block of data is encrypted with it's own random symmetric key.  This key is then encrypted using the given public key and stored with the data.

The basic concept of pgp is show in this image:

![](images/PGP-diagram-wikipedia-479x500.jpg?raw=true)

_Image taken from wikipedia_

##REST API

The REST API interface can be used to perform operations against the Filesystem.  Default location is localhost:8010.

[Apiary.io Documenation](http://docs.blockerapi.apiary.io)

##Example code
[Example test scenario](https://github.com/Inflatablewoman/blocker/blob/master/server/server_test.go)
