#!/usr/bin/env bash
docker pull cloudurable/cassandra-cloud:latest
docker run  -it --name runner-cc  \
-p 80:80 \
-v `pwd`:/gopath/src/github.com/cloudurable/cassandra-cloud \
cloudurable/cassandra-cloud:latest
docker rm runner-cc
