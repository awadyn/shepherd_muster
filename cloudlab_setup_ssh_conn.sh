# get node ip list as input
# for every node, gen ssh key and copy its pub key to all other nodes
node_list=( "$@" )
myself="awadyn"
for i in "${!node_list[@]}"; do
	echo "Setup for ${node_list[i]}";
	# skip steps if connection is already established
	ssh $myself@${node_list[i]} 'if [ -f ".ssh/id_rsa.pub" ]; then echo "id_rsa.pub exists.. skipping ssh-keygen"; else ssh-keygen; fi';
	id_rsa=$(ssh $myself@${node_list[i]} 'cat .ssh/id_rsa.pub')
	ip_1=$(ssh $myself@${node_list[i]} 'ifconfig | grep -B1 10.10.1 | grep inet | grep -oP "inet \K(\d+\.\d+\.\d+\.\d+)"')
	ip_2=$(ssh $myself@${node_list[i]} 'ifconfig | grep -B1 128.110.96 | grep inet | grep -oP "inet \K(\d+\.\d+\.\d+\.\d+)"')
	#echo "$id_rsa";
	for j in "${!node_list[@]}"; do
		if [[ $i -eq $j  ]]; then 
			continue; 
		else
			echo "Adding ${node_list[i]} to authorized_keys of ${node_list[j]}: ";
			ssh $myself@${node_list[j]} "if grep -q '$id_rsa' .ssh/authorized_keys; then echo '${node_list[i]} is known to this host'; else echo '$id_rsa' >> .ssh/authorized_keys; fi";
			echo "Adding ${node_list[i]} to known_hosts of ${node_list[j]}: ";
			ssh $myself@${node_list[j]} "ssh-keyscan -H ${node_list[i]} >> .ssh/known_hosts; ssh-keyscan -H $ip_1 >> .ssh/known_hosts; ssh-keyscan -H $ip_2 >> .ssh/known_hosts";
		fi
	done
	echo "";
done

echo "Establishing first connection:"
for i in "${!node_list[@]}"; do
	for j in "${!node_list[@]}"; do
		ip_1=$(ssh $myself@${node_list[j]} 'ifconfig | grep -B1 10.10.1 | grep inet | grep -oP "inet \K(\d+\.\d+\.\d+\.\d+)"')
		ip_2=$(ssh $myself@${node_list[j]} 'ifconfig | grep -B1 128.110.96 | grep inet | grep -oP "inet \K(\d+\.\d+\.\d+\.\d+)"')
		if [[ $i -eq $j  ]]; then 
			continue;
		else
			ssh $myself@${node_list[i]} "ssh $myself@${node_list[j]} 'exit'";
			echo "${node_list[i]} -> ${node_list[j]} .. OK"
			ssh $myself@${node_list[i]} "ssh $myself@$ip_1 'exit'";
			echo "${node_list[i]} -> $ip_1 .. OK"
			ssh $myself@${node_list[i]} "ssh $myself@$ip_2 'exit'";
			echo "${node_list[i]} -> $ip_2 .. OK"
		fi 
	done
	echo ""
done

