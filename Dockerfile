FROM       arken/gom-base
MAINTAINER Damien Metzler <dmetzler@nuxeo.com>
RUN mkdir -p /opt/arken
ADD . /opt/arken/
COPY arken-docker.yml /opt/arken/arken.yml
WORKDIR /opt/arken
RUN gom install
#RUN gom build
#ENTRYPOINT ["/opt/arken/arken", "serve"]
