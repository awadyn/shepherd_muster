node_list=( "$@" )
myself="awadyn"
go_version="1.22.4"
for node in ${node_list[*]}; do
	echo "Setting up golang on $node";
	ssh $myself@$node "if [[ $(which go) ]]; 
		then echo 'go$go_version found at $(which go)'; 
		else echo 'golang not found.. Downloading.. $go_version'; 
		wget https://go.dev/dl/go$go_version.linux-amd64.tar.gz; 
		sudo tar -C /usr/local -xzf go$go_version.linux-amd64.tar.gz; 
		echo 'export PATH=$PATH:$(which go)' >> .bashrc;
		fi"

	echo "Setting up shepherd_muster on $node";
	ssh $myself@$node "if ! [ -d 'shepherd_muster' ]; 
		then git clone https://github.com/awadyn/shepherd_muster.git; 
		fi"
done
