taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value
echo > all_mutilate_output.txt
ssh -f 130.127.133.42 "./read_rapl_start.sh"

for qps in 400000 900000 900000 600000 200000 400000 600000 100000 600000 200000 100000 400000 900000; do
	echo $qps;
	if [[ $qps == "100000" ]]; then
		ssh 130.127.133.42 "sudo wrmsr -a 0x199 0x1a00; sudo ethtool -C enp130s0f0 rx-usecs 300";
	elif [[ $qps == "200000" ]]; then
		ssh 130.127.133.42 "sudo wrmsr -a 0x199 0xc00; sudo ethtool -C enp130s0f0 rx-usecs 300";
	elif [[ $qps == "400000" ]]; then
		ssh 130.127.133.42 "sudo wrmsr -a 0x199 0x1a00; sudo ethtool -C enp130s0f0 rx-usecs 300";
	elif [[ $qps == "600000" ]]; then
		ssh 130.127.133.42 "sudo wrmsr -a 0x199 0x1a00; sudo ethtool -C enp130s0f0 rx-usecs 200";
	elif [[ $qps == "900000" ]]; then
		ssh 130.127.133.42 "sudo wrmsr -a 0x199 0x1100; sudo ethtool -C enp130s0f0 rx-usecs 100";
	fi

	taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --measure_connections=512 --measure_qps=2000 --qps=$qps --time=30 >> all_mutilate_output.txt

	ssh -f awadyn@130.127.133.42 "./read_rapl.sh"
done
