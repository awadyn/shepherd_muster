#!/bin/bash

mcd_server=$1
iface=$2
dir=$3
qps_=$4

for qps in $qps_; do
#for qps in 100000 200000 400000 600000 800000 1000000 1200000 1400000 1600000; do
	for itrd in 1 2 20 50 100 150 200 250 300 350 400; do 
                echo "------------- ITRD: $itrd -- QPS: $qps";

		subdir=$dir/qps_$qps/itrd_$itrd/
		mkdir -p $subdir

		# start memcached server
                ssh $mcd_server "taskset -c 0-6,8-14 ~/memcached/memcached -u nobody -t 14 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_dump 2>&1 < /dev/null &"
		sleep 1;
		# preload memcached database
		taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value

		# start mutilate agents
                for i in 3 4 5; do
                        ssh 10.10.1.$i "~/mutilate/mutilate --agentmode --threads=16 > mutilate_dump 2>&1 < /dev/null &";
                done

		# set itrd and flush current ixgbe logs
		ssh $mcd_server "sudo ethtool -C $iface rx-usecs $itrd; ./flush_ixgbe_logs.sh &> /dev/null; sudo ./read_rapl_start.sh";
		
		# start mutilate run
		taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.5} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=$qps --time=30; 
		
		# get rapl readings for this run
                ssh $mcd_server "sudo ./read_rapl.sh; ./cat_ixgbe_logs.sh ixgbe_logs/; sudo killall memcached";
                for i in 3 4 5; do
                        ssh 10.10.1.$i "sudo killall mutilate";
                done

		# scp ixgbe logs
                echo;
		scp -r $mcd_server:~/ixgbe_logs/* $dir/qps_$qps/itrd_$itrd/
                scp -r $mcd_server:~/rapl_* $dir/qps_$qps/itrd_$itrd/;
		echo;
                ./compute_rapl_joules.sh $dir/qps_$qps/itrd_$itrd/rapl_pkg.txt;
                ./compute_rapl_joules.sh $dir/qps_$qps/itrd_$itrd/rapl_pp0.txt;
                echo; echo "--------------------------------"; echo;
#		break
	done
done





