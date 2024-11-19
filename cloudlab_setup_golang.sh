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
fi

echo "Installing protobuf grpc compiler.."
sleep 1
sudo apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"

# use protobuf-compiler
# protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto-dir/file.proto
# python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. shep_optimizer/shep_optimizer.proto
