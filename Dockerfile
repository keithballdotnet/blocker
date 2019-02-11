# The following commands can be used to create, run and publish a docker image
# Ensure boot2Docker is running :  boot2docker up
# Command Used to locate docker in shell : export DOCKER_HOST=tcp://BLAH BLAH 
# Command Used to build image : docker build -t keithballdotnet/blocker .
# Command Used to publish image : docker push keithballdotnet/blocker
# Command Used to run image : docker run -e "CB_HOST=http://COUCHBASE_ADDRESS:8091" --publish 6060:8002 --name blocker1 --rm keithballdotnet/blocker

# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/keithballdotnet/blocker
#ADD /Users/keithball/Projects/blocker/src/github.com/couchbaselabs /go/src/github.com/couchbaselabs

# Build the blocker command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get "code.google.com/p/go-uuid/uuid"
RUN go get "code.google.com/p/snappy-go/snappy"
RUN go get "github.com/couchbaselabs/go-couchbase"
RUN go get "github.com/rcrowley/go-tigertonic"
RUN go get "gopkg.in/check.v1"
RUN go install github.com/keithballdotnet/blocker
RUN mkdir /tmp/blocks/

# Run the blocker command by default when the container starts.
ENTRYPOINT /go/bin/blocker

# Document that the service listens on port 8002.
EXPOSE 8002