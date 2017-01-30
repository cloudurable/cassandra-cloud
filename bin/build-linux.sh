#!/usr/bin/env bash


rm cassandra-cloud

set -e

cd /gopath/src/github.com/cloudurable/cassandra-cloud/
source ~/.bash_profile
export GOPATH=/gopath


echo "Running go clean"
go clean
echo "Running go get"
go get
echo "Running go build"
go build

ls


