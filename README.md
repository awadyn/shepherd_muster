# MustHerd

## Preparing MustHerd Environment
#### Assuming one shepherd node and at least one muster node:
ssh user@node-shepherd 'sudo apt update && sudo apt upgrade -y'
ssh user@node-muster-x 'sudo apt update && sudo apt upgrade -y'

#### Clone MustHerd code base on shepherd node and all muster nodes:
user@node:$ git clone https://github.com/awadyn/shepherd_muster.git
user@node:$ cd shepherd_muster; ./cloudlab_setup_golang.sh

#### Running above script checks for a compatible golang version:
user@node:$ go_version=1.22.4							// compatible golang version
user@node:$ which go								// checks if go runtime is installed
user@node:$ go version 								// checks go version
user@node:$ sudo rm -rf /usr/local/bin/go 					// remove current go version
user@node:$ wget https://go.dev/dl/go$go_version.linux-amd64.tar.gz		// download go version
user@node:$ sudo tar -C /usr/local -xzf go$go_version.linux-amd64.tar.gz	// install go locally
user@node:$ echo 'export PATH=$PATH:/usr/local/go/bin' >> .bashrc		// add go binary to bash shell environment
user@node:$ export PATH=$PATH:/usr/local/go/bin					// add go binary to bash shell path

## Running MustHerd Test
