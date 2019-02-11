Blocker
=======
A block based filesystem microservice written in go

[![GoDoc](https://godoc.org/github.com/Inflatablewoman/blocker?status.svg)](https://godoc.org/github.com/Inflatablewoman/blocker)
[![Build Status](https://travis-ci.org/Inflatablewoman/blocker.svg)](https://travis-ci.org/Inflatablewoman/blocker)
[![Coverage Status](https://coveralls.io/repos/Inflatablewoman/blocker/badge.svg)](https://coveralls.io/r/Inflatablewoman/blocker)

## The case for Blocker

### What is blocker?

Blocker is a block based file storage microservice. Files are stored in blocks referenced by their hash value. Blocks are encrypted. Blocks are compressed. Blocker is storage provider agnostic. Blocker can save storage space by removing duplication of data.

Imagine the following upload of a 14MB video file. The next time this document is uploaded or copied in the system then blocker will check for the existence of the hashes h1, h2, h3 and h4. If the hashes exist then no action will be taken.

![](images/DropboxFileFormat.png?raw=true)

_Image taken from blogs.dropbox.com_

If a change is made to the video file, for example an extra 1mb of data is appended and uploaded. Then the document will now exist as h1, h2, h3 and the newly created h5. h4 will remain stored and be returned if the old version of the document is requested.

The basis of this is taken from a [2014 tech blog from Dropbox](https://blogs.dropbox.com/tech/2014/07/streaming-file-synchronization/).

## Features

- Files are stored in blocks
- Immutable blocks (Append Only Data Store)
- Blocks are hashed
- Blocks are encrypted
- Blocks are compressed
- Reduces data duplication
- Files where blocks have not changed reference old blocks
- A REST interface for manipulating blocks
- Uses couchbase for BlockedFiles repository
- Possible to specify crypto provider
   + openpgp - Encrypt using pgp key pair
   + aws - Use keys retrieved from AWS KMS
   + gokms - Use keys retrieved from GO-KMS
- Possible to specify a backend Storage provider
   + nfs - Local mount disk storage (GlusterFS could be used)
   + couchbase - Couchbase Raw Binary storage
   + azure - Azure Simple Storage
   + s3 - Amazon s3 storage

## Todo

- Updating a block from the block list

## Authorization

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

## Compression

Compression is done using google's [Snappy compression](https://code.google.com/p/snappy/).

## Data Encryption

Data encryption can be done using one of either the following providers.  You can select which mode by setting the cli flag *-c* to either *"go-kms"*, *"openpgp"* or *"aws"*.  OpenPGP is the default crypto provider.

### GO Key Management Service

GO-KMS can is a Key Management Service written in GO.  It is available on [github.com](https://github.com/Inflatablewoman/go-kms).  GO-KMS is AWK KMS compatible.

To setup blocker to run with GO-KMS the following should be set.

```
export BLOCKER_GOKMS_AUTHKEY=YourGoKmsKey
export BLOCKER_GOKMS_URL=https://go-kms.yourhost.com

#optional: export BLOCKER_GOKMS_KEYID=YourKeyID
#If left empty the first availble key will be selected
```

The crypto provider uses [AES](http://en.wikipedia.org/wiki/Advanced_Encryption_Standard) and a key size of 256bits using the [GCM cipher](http://en.wikipedia.org/wiki/Galois/Counter_Mode) to provide confidentiality as well as authentication.  

### AWS Key Management Service

Communication with the Amazon Web Services (AWS) is done via the [official library]("http://www.github.com/awslabs/aws-sdk-go/aws").

To setup blocker to run with AWS the following should be set.

```
export BLOCKER_KMS_KEY=YourAwsKey
export BLOCKER_KMS_SECRET=YourAwsSecret

#optional: export BLOCKER_KMS_REGION=eu-central-1
#Default value is: eu-central-1

#optional: export BLOCKER_KMS_KEY_ID=YourKeyID
#If left empty the first availble key from the region will be selected

```

Key management in KMS is very simple.  Refer to the AWS documentation for more information on how to set up an encryption key.  Here I have created a key for blocker.

![](images/aws_key_management.png?raw=true)

I can use AWS Identity and Access Management (IAM) to create a user that can access the key, with the required authorization and define a policy that will only allow access from certain IP addresses.

The encryption follows the pattern as specified in the in the [KMS Cryptographic Whitepaper](https://d0.awsstatic.com/whitepapers/KMS-Cryptographic-Details.pdf).

For each block a new DataKey will be requested from KMS.  The key will return an encrypted version of the key and a plaintext version of the key.  The plaintext version of the key will be used to encrypt the data.  It will be then combined into an envelop of data ready for persistence.

![](images/aws_encrypt.png?raw=true)

Upon a request for decryption the data envelope will be inspected, the encrypted key extracted and then decrypted by the KMS server.  The decrypted key can then be used to decrypt the body of the data.

![](images/aws_decrypt.png?raw=true)

### OpenPGP

The OpenPGP crypto provider is done using using the [openpgp golang library](https://godoc.org/golang.org/x/crypto/openpgp).  Specifically a SHA256 hash and the AES256 Cipher is used for encryption.  

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

## REST API

The REST API interface can be used to perform operations against the Filesystem.  Default location is localhost:8010.

[Apiary.io Documenation](http://docs.blockerapi.apiary.io)

##Example code
[Example test scenario](https://github.com/Inflatablewoman/blocker/blob/master/server/server_test.go)
