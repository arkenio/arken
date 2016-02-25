FROM       arken/gom-base
MAINTAINER Damien Metzler <dmetzler@nuxeo.com>
RUN go get github.com/arkenio/arken
WORKDIR /usr/local/go/src/github.com/arkenio/arken
RUN gom install
RUN gom test

ENTRYPOINT ["arken", "--etcdAddress", "http://172.17.42.1:4001", "serve"]
