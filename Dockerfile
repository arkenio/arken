FROM       arken/gom-base
MAINTAINER Damien Metzler <dmetzler@nuxeo.com>
ADD . /opt/arken/
WORKDIR /opt/arken
RUN gom install
RUN gom build
ENTRYPOINT ["/opt/arken/arken","--config","arken-docker.yml", "serve"]
