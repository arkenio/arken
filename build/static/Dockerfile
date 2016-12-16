FROM nuxeo:{{.nuxeoversion}}

RUN $NUXEO_HOME/bin/nuxeoctl mp-init \
 && $NUXEO_HOME/bin/nuxeoctl mp-install {{.nuxeopackages}}

ENTRYPOINT ["/docker-entrypoint.sh"]
EXPOSE 8080
EXPOSE 8787
CMD ["nuxeoctl","console"]