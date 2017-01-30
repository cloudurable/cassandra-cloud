#!/bin/bash
set -e

yum -y install wget
yum install -y git
yum install -y gcc
yum install -y systemd-devel
yum install -y nano

echo "installing go"
cd /tmp
wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
tar -C /usr/local/ -xzf go1.7.3.linux-amd64.tar.gz
rm go1.7.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile


wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
chmod +x jq-linux64
mv jq-linux64 /usr/bin/jq


