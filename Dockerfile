FROM ubuntu:14.04
MAINTAINER <EMC{code}>

# To build this dockerfile first ensure that it is named "Dockerfile"
# make sure that a directory "docker_resources" is also present in the same directory as "Dockerfile",
#   and that "docker_resources" contains the files "go-wrapper" and "get_go-bindata_md5.sh"

# Assuming:
# Your dockerhub username: dhjohndoe
# Your github username: ghjohndoe
# Your REX-Ray fork is checked out in $HOME/go/src/github.com/ghjohndoe/libstorage/

# To build a Docker image using this Dockerfile:
# docker build -t dhjohndoe/golang-glide:0.1.0 .

# To build libstorage using this Docker image:
# docker pull dhjohndoe/golang-glide:0.1.0
# If cutting and pasting the next line remove '\' and '#' characters remember to replace ghjohndoe and dhjohndoe
# docker run -v $HOME/go/src/github.com/ghjohndoe/libstorage/:/go/src/github.com/emccode/libstorage/ \
# -v $HOME/build/libstorage/pkg/:/go/pkg/ \
# -v $HOME/build/libstorage/bin/:/go/bin/ \
# -w=/go/src/github.com/emccode/libstorage/ dhjohndoe/golang-glide:0.1.0

# after building build resources will be placed in $HOME/build/libstorage/ in the pkg/ and bin/ directories

RUN apt-get update && apt-get install -y --no-install-recommends software-properties-common
RUN add-apt-repository ppa:masterminds/glide

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        g++ \
        gcc \
        git \
        glide \
        make \
    && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.6.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 e40c36ae71756198478624ed1bb4ce17597b3c19d243f3f0899bb5740d56212a

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
    && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -xzf golang.tar.gz \
    && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

COPY docker_resources/go-wrapper /usr/local/bin/

VOLUME ["/go/src/github.com/emccode/libstorage", "/go/pkg", "/go/bin"]

CMD ["/bin/bash", "-c", "make clean && make version && make deps && make build"]

# manual build steps:
# make clean
# make version
# make deps
# make build
# The libstorage build output will be in: /go/bin/ /go/pkg/
