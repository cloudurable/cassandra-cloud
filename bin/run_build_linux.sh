#!/usr/bin/env bash
docker pull cloudurable/cassandra-cloud:latest
docker run -it --name build -v `pwd`:/gopath/src/github.com/cloudurable/cassandra-cloud \
cloudurable/cassandra-cloud \
/bin/sh -c "/gopath/src/github.com/cloudurable/cassandra-cloud/bin/build-linux.sh"
docker rm build
mv cassandra-cloud cassandra-cloud_linux

