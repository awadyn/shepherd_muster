#!/bin/bash

go_version="1.22.4"
bin_dir=/usr/local

echo "Setting up go$go_version:"
sleep 1
if [[ $(which go) ]]; 
then echo "golang found at:"; which go; go version; sleep 1;
else echo "golang not found.. Downloading go$go_version:"; sleep 1; 
	wget https://go.dev/dl/go$go_version.linux-amd64.tar.gz; 
	sudo tar -C $bin_dir -xzf go$go_version.linux-amd64.tar.gz;	
	echo "export PATH=$PATH:$bin_dir/go/bin" >> ~/.bashrc;
	which go; go version;
fi

