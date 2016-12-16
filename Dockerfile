FROM golang:1.7
MAINTAINER      Damien METZLER <dmetzler@nuxeo.com>

RUN go get github.com/tools/godep
RUN go get github.com/mjibson/esc
RUN CGO_ENABLED=0 go install -a std
ENV APP_DIR $GOPATH/src/github.com/arkenio/arken

# Set the entrypoint as the binary, so `docker run <image>` will behave as the binary
ENTRYPOINT      ["/arken"]
ADD arken-docker.yml /arken.yml
ADD . $APP_DIR
# Compile the binary and statically link
RUN cd $APP_DIR && \
    esc -o api/static.go -prefix static -pkg api static && \
    esc -o build/static.go -prefix build/static -pkg build build/static && \
    CGO_ENABLED=0 godep restore && \
    rm -rf $GOPATH/src/github.com/rancher/go-rancher/vendor/github.com/gorilla/websocket
RUN cd $APP_DIR && godep go build -o /arken -ldflags '-w -s'
CMD serve
