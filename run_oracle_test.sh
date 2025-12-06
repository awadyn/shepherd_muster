mcd_server=$1
iface=$2

echo > ondemand_dvfs_lat_joules_14_cores_c6220.txt

ssh $mcd_server "taskset -c 0-6,8-14 ~/memcached/memcached -u nobody -t 14 -m 32G -c 8192 -b 8192 -l $mcd_server -B binary > mcd_dump 2>&1 < /dev/null &"
for i in 3 4 5; do
	ssh 10.10.1.$i "~/mutilate/mutilate --agentmode --threads=16 > mutilate_dump 2>&1 < /dev/null &";
done

taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value
sleep 1
echo "STARTING.."

ssh $mcd_server "sudo ./read_rapl_start_c6220.sh"
#ssh $mcd_server "sudo ethtool -C $iface rx-usecs 1";

for qps in 400000 800000 600000 200000 400000 100000 600000 200000 100000 800000; do
	ssh $mcd_server "./flush_ixgbe_logs_c6220.sh &> /dev/null";
			
	if [[ $qps == "100000" ]]; then
		ssh $mcd_server "sudo ethtool -C $iface rx-usecs 300";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0xc00";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0xc00; sudo ethtool -C $iface rx-usecs 300";
	elif [[ $qps == "200000" ]]; then
		ssh $mcd_server "sudo ethtool -C $iface rx-usecs 300";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0xc00";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0xc00; sudo ethtool -C $iface rx-usecs 300";
	elif [[ $qps == "400000" ]]; then
		ssh $mcd_server "sudo ethtool -C $iface rx-usecs 250";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0xc00";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0xc00; sudo ethtool -C $iface rx-usecs 250";
	elif [[ $qps == "600000" ]]; then
		ssh $mcd_server "sudo ethtool -C $iface rx-usecs 250";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0x1100";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0x1100; sudo ethtool -C $iface rx-usecs 250";
	elif [[ $qps == "800000" ]]; then
		ssh $mcd_server "sudo ethtool -C $iface rx-usecs 200";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0x1100";
		#ssh $mcd_server "sudo wrmsr -a 0x199 0x1100; sudo ethtool -C $iface rx-usecs 200";
	fi

	taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.5} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_qps=2000 --qps=$qps --time=30 >> ondemand_dvfs_lat_joules_14_cores_c6220.txt

	ssh $mcd_server "sudo ./read_rapl_c6220.sh";
done

for i in 3 4 5; do
        ssh 10.10.1.$i "sudo killall mutilate";
done
ssh $mcd_server "sudo killall memcached"

scp -r $mcd_server:~/rapl_* .;

