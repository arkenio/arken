

arken
========

[![Build Status](https://travis-ci.org/arkenio/arken.png?branch=master)](https://travis-ci.org/arkenio/arken)

Arken is a daemon tool that takes care of an Arken cluster and exposes an API on top to that.

## Usage

### Starting

You can launch arken by with the following command

	# arken serve

### Rest API

Two endpoints provides some information on Arken.

    GET http://localhost:8888/api/v1/services
    GET http://localhost:8888/api/v1/services/{serviceId}
    PUT http://localhost:8888/api/v1/services/{serviceId}?action=start
    PUT http://localhost:8888/api/v1/services/{serviceId}?action=stop
    PUT http://localhost:8888/api/v1/services/{serviceId}?action=passivate

    GET http://localhost:8888/api/v1/domains/
    GET http://localhost:8888/api/v1/domain/{domainName}

For complete API documentation go to the doc page : http://localhost:8888/doc/


### Web Socket 

A websocket is available at ws://localhost:8888/ws/ where ModelEvent are pushed.


### Authentication

By adding some keys in the `arken.yml` configuration file, you can add your some
credentials like that :


    apiKeys:
      akey:
        accessKey: A23DR
        secretKey: secret



## Report & Contribute


We are glad to welcome new developers on this initiative, and even simple usage feedback is great.
- Ask your questions on [Nuxeo Answers](http://answers.nuxeo.com)
- Report issues on this github repository (see [issues link](http://github.com/arkenio/arken/issues) on the right)
- Contribute: Send pull requests!


## About Nuxeo

Nuxeo provides a modular, extensible Java-based
[open source software platform for enterprise content management](http://www.nuxeo.com/en/products/ep),
and packaged applications for [document management](http://www.nuxeo.com/en/products/document-management),
[digital asset management](http://www.nuxeo.com/en/products/dam) and
[case management](http://www.nuxeo.com/en/products/case-management).

Designed by developers for developers, the Nuxeo platform offers a modern
architecture, a powerful plug-in model and extensive packaging
capabilities for building content applications.

More information on: <http://www.nuxeo.com/>
