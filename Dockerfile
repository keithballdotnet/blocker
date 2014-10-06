# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/Inflatablewoman/blocker
#ADD /Users/keithball/Projects/blocker/src/github.com/couchbaselabs /go/src/github.com/couchbaselabs

# Build the blocker command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get "code.google.com/p/go-uuid/uuid"
RUN go get "code.google.com/p/snappy-go/snappy"
RUN go get "github.com/Inflatablewoman/go-couchbase"
RUN go get "github.com/rcrowley/go-tigertonic"
RUN go get "gopkg.in/check.v1"
RUN go install github.com/Inflatablewoman/blocker
RUN mkdir /tmp/blocks/
RUN go test -v github.com/Inflatablewoman/blocker/crypto

# Run the blocker command by default when the container starts.
ENTRYPOINT /go/bin/blocker

# Document that the service listens on port 8002.
EXPOSE 8002