#!/bin/bash

mcd_server=$1
iface=$2
dir=$3
qps_=$4

#for qps in 100000 200000 400000 600000 800000 1000000 1200000 1400000 1600000; do
for qps in $qps_; do
	mkdir -p $dir/qps_$qps/
	for itrd in 1 2 20 50 100 150 200 250 300 350 400; do 
		for dvfs in 0xc00 0x1100 0x1500 0x1900 0x1c00; do 
	                echo "------------- ITRD: $itrd -- QPS: $qps -- DVFS: $dvfs";
	
			subdir=$dir/qps_$qps/itrd_$itrd\_dvfs_$dvfs/
			mkdir -p $subdir
	
			# start memcached server
	                ssh $mcd_server "taskset -c 0-8,10-18 ~/memcached/memcached -u nobody -t 18 -m 32G -c 8192 -b 8192 -l $mcd_server -B binary > mcd_dump 2>&1 < /dev/null &"
			sleep 1;
			# preload memcached database
			taskset -c 0 ~/mutilate/mutilate --binary -s $mcd_server --loadonly -K fb_key -V fb_value
	
			# start mutilate agents
	                for i in 3 4 5 6; do
	                        ssh 10.10.1.$i "~/mutilate/mutilate --agentmode --threads=16 > mutilate_dump 2>&1 < /dev/null &";
	                done
	
			# set itrd and flush current ixgbe logs
			ssh $mcd_server "mkdir -p ~/ixgbe_logs/; sudo ethtool -C $iface rx-usecs $itrd; sudo wrmsr -a 0x199 $dvfs; ./flush_ixgbe_logs.sh &> /dev/null; sudo ./read_rapl_start_c8220.sh";
			
			# start mutilate run
			taskset -c 0 ~/mutilate/mutilate --binary -s $mcd_server --noload --agent={10.10.1.3,10.10.1.4,10.10.1.5,10.10.1.6} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=$qps --time=30; 
			
			# get rapl readings for this run
	                ssh $mcd_server "sudo ./read_rapl_c8220.sh; ./cat_ixgbe_logs.sh ixgbe_logs/; sudo killall memcached";
	                for i in 3 4 5 6; do
	                        ssh 10.10.1.$i "sudo killall mutilate";
	                done
	
			# scp ixgbe logs
	                echo;
			scp -r $mcd_server:~/ixgbe_logs/* $dir/qps_$qps/itrd_$itrd\_dvfs_$dvfs/
	                scp -r $mcd_server:~/rapl_* $dir/qps_$qps/itrd_$itrd\_dvfs_$dvfs/;
			echo;
	                ./compute_rapl_joules.sh $dir/qps_$qps/itrd_$itrd\_dvfs_$dvfs/rapl_pkg.txt;
	                ./compute_rapl_joules.sh $dir/qps_$qps/itrd_$itrd\_dvfs_$dvfs/rapl_pp0.txt;
	                echo; echo "--------------------------------"; echo;
			#break
		done
		#break
	done
	#break
done





