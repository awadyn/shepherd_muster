#!/bin/bash
ssh 10.10.1.3 "sudo killall mutilate"
ssh 10.10.1.4 "sudo killall mutilate"
ssh 10.10.1.6 "sudo killall mutilate"
ssh 10.10.1.5 "sudo killall mutilate"
ssh 10.10.1.2 "sudo killall memcached"

sleep 1 

echo "START MEMCACHED"
## start one mcd proc on node4
#ssh 10.10.1.5 "taskset -c 0-15 ~/memcached/memcached -p 11211 -u nobody -t 16 -m 32G -c 8192 -b 8192 -l 10.10.1.5 -B binary > mcd_11211.log 2>&1 < /dev/null &"
# start one mcd proc on node1
ssh 10.10.1.2 "taskset -c 0-15 ~/memcached/memcached -p 11211 -u nobody -t 16 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_11211.log 2>&1 < /dev/null &"

#ssh 10.10.1.2 "taskset -c 1-7,9-15 ~/memcached/memcached -p 11211 -u nobody -t 14 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_11211.log 2>&1 < /dev/null &"
#ssh 10.10.1.2 "taskset -c 1-3,9-12 ~/memcached/memcached -p 11211 -u nobody -t 7 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_11211.log 2>&1 < /dev/null &"
#ssh 10.10.1.2 "taskset -c 4-7,13-15 ~/memcached/memcached -p 11212 -u nobody -t 7 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_11212.log 2>&1 < /dev/null  &"


echo "START LOAD GENERATION AGENTS"
ssh 10.10.1.5 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.4 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.6 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
sleep 1

echo "LOAD MEMCACHED DATABASE"
#taskset -c 0 ~/mutilate/mutilate -vv --binary -s 10.10.1.5:11211 --loadonly -K fb_key -V fb_value
taskset -c 0 ~/mutilate/mutilate -vv --binary -s 10.10.1.2:11211 --loadonly -K fb_key -V fb_value
#taskset -c 0 ~/mutilate/mutilate -vv --binary -s 10.10.1.2:11212 --loadonly -K fb_key -V fb_value

mkdir -p exp/
rm -rf exp/*
echo > exp/mcd_11211.out
#echo > exp/mcd_11212.out

sleep 1

echo "FLUSH IXGBE LOGS"
ssh 10.10.1.2 "mkdir -p ixgbe_logs/; cat /proc/ixgbe_stats/core/* | wc -l"

echo "START EXPERIMENT"
#for qps1 in 100000 300000 500000 700000 900000 1100000 1300000 1500000 1700000 1900000 2100000 2300000; do 
#for qps1 in 100000 300000 500000 600000 700000 800000 900000 1000000 1100000 1200000 1400000 1600000; do 
for qps1 in 100000 200000 300000 400000 500000 600000 700000 800000 900000 1000000 1100000 1200000; do 
#	for qps2 in 200000 400000 800000 1200000 1600000; do

#		echo "START MUSTHERD"
#		ssh 10.10.1.2 "cd ~/shepherd_muster/shep_remote_muster/; taskset -c 0,8 /usr/local/go/bin/go run remote_muster/* 10.10.1.2 10.10.1.1 16 50051 50061 50071 50081 > muster.log 2>&1 < /dev/null &"
#		sleep 3 && cd ~/shepherd_muster/shep_remote_muster/;
#		taskset -c 1-7,9-15 /usr/local/go/bin/go run shepherd/* >> shepherd.log &
#		sleep 3 && cd ~/;
		
#		mkdir -p exp/$qps1\_$qps2/
#		echo $qps1 $qps2
		mkdir -p exp/$qps1/
		echo $qps1

		ssh 10.10.1.2 "cat /proc/ixgbe_stats/core/* | wc -l"

		# read rapl
		ssh 10.10.1.2 "./read_rapl_start_c6220.sh"

#		# start load generation for 1 mcd process on node4
#		taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.5:11211 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.6} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.75 --qps=$qps1 --depth=128 --measure_connections=64 --measure_qps=2000 --time=30 >> exp/mcd_11211.out &
		# start load generation for 1 mcd process on node1
		taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2:11211 --noload --agent={10.10.1.4,10.10.1.6,10.10.1.5} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.75 --qps=$qps1 --depth=128 --measure_connections=64 --measure_qps=2000 --time=30 >> exp/mcd_11211.out &

#		taskset -c 8 ~/mutilate/mutilate --binary -s 10.10.1.2:11212 --noload --agent=10.10.1.4 --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --qps=$qps2 --depth=64 --measure_connections=32 --measure_qps=2000 --time=30 >> exp/mcd_11212.out &
		
		# sleep and wait..
		sleep 33
		
		# read rapl
		ssh 10.10.1.2 "./read_rapl_c6220.sh"

#		ssh 10.10.1.2 "sudo killall go; sudo killall read_ixgbe_stats; sudo killall bayopt_main"
#		sudo killall go; sudo killall defs
#		sleep 1

#		cp ~/shepherd_muster/shep_remote_muster/mustherd-logs-muster-10.10.1.2/* exp/$qps1\_$qps2/
#		cp ~/shepherd_muster/shep_remote_muster/mustherd-logs-muster-10.10.1.2/* exp/$qps1/

		ssh 10.10.1.2 'for i in {0..15}; do cat /proc/ixgbe_stats/core/$i > ixgbe_logs/core$i; done'
#		scp -r 10.10.1.2:~/ixgbe_logs/* exp/$qps1\_$qps2/
		scp -r 10.10.1.2:~/ixgbe_logs/* exp/$qps1/

		scp -r 10.10.1.2:~/rapl_* exp/$qps1/
#	done
done

ssh 10.10.1.3 "sudo killall mutilate"
ssh 10.10.1.4 "sudo killall mutilate"
ssh 10.10.1.6 "sudo killall mutilate"
ssh 10.10.1.2 "sudo killall memcached"
ssh 10.10.1.5 "sudo killall mutilate"
