FROM golang:1.5 
MAINTAINER Damien Metzler <dmetzler@nuxeo.com>
RUN go get github.com/mattn/gom
ADD Gomfile /opt/arken/Gomfile
WORKDIR /opt/arken
RUN gom install
ADD . /opt/arken/
RUN gom build
ENTRYPOINT ["/opt/arken/arken","--config","arken-docker.yml", "serve"]
