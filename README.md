# MustHerd

## Preparing MustHerd Environment
#### Assuming one shepherd node and at least one muster node:
ssh username@node-shepherd 'sudo apt update && sudo apt upgrade -y'
ssh username@node-muster-x 'sudo apt update && sudo apt upgrade -y'

#### Check for a compatible golang version on shepherd node and all muster nodes:
username@node:$ go_version=1.22.4
username@node:$ which go			// checks if go runtime is installed
username@node:$ go --version 			// checks go version
username@node:$ sudo rm -rf /usr/local/bin/go 	// remove current go version
username@node:$ wget https://go.dev/dl/go$go_version.linux-amd64.tar.gz
username@node:$ sudo tar -C /usr/local -xzf go$go_version.linux-amd64.tar.gz
username@node:$ echo 'export PATH=$PATH:/usr/local/go/bin' >> .bashrc 
username@node:$ export PATH=$PATH:/usr/local/go/bin

#### Clone MustHerd code base on shepherd node and all muster nodes:
username@node:$ git clone https://github.com/awadyn/shepherd_muster.git

## Running MustHerd Test
