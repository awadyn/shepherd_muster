#!/bin/bash

echo "START MEMCACHED"
# start one mcd proc on node1
ssh 10.10.1.2 "taskset -c 1-7,9-15 ~/memcached/memcached -p 11211 -u nobody -t 14 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_11211.log 2>&1 < /dev/null &"

echo "START LOAD GENERATION AGENTS"
ssh 10.10.1.3 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.4 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.6 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"

sleep 1

echo "LOAD MEMCACHED DATABASE"
taskset -c 0 ~/mutilate/mutilate -vv --binary -s 10.10.1.2:11211 --loadonly -K fb_key -V fb_value

mkdir -p exp/
rm -rf exp/*
echo > exp/mcd_11211.out

sleep 1

echo "START EXPERIMENT"

# read rapl
ssh 10.10.1.2 "./read_rapl_start_c6220.sh"

echo "FLUSH IXGBE LOGS"
ssh 10.10.1.2 "mkdir -p ixgbe_logs/; cat /proc/ixgbe_stats/core/* | wc -l"

#for qps1 in 300000 600000 900000 1200000 1500000 1800000; do 
for qps1 in 100000 200000 300000 400000 600000 800000; do 
	# start load generation for 1 mcd process on node1
	taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2:11211 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.6} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.75 --qps=$qps1 --depth=128 --measure_connections=64 --measure_qps=2000 --time=20 >> exp/mcd_11211.out &

	sleep 23

	# read rapl
	ssh 10.10.1.2 "./read_rapl_c6220.sh"

done

echo "SAVE IXGBE LOGS"
ssh 10.10.1.2 'for i in {0..15}; do cat /proc/ixgbe_stats/core/$i > ixgbe_logs/core$i; done'
scp -r 10.10.1.2:~/ixgbe_logs/* exp/
scp -r 10.10.1.2:~/rapl_* exp/

echo "KILLING SERVER AND AGENTS"
ssh 10.10.1.3 "sudo killall mutilate"
ssh 10.10.1.4 "sudo killall mutilate"
ssh 10.10.1.6 "sudo killall mutilate"
ssh 10.10.1.2 "sudo killall memcached"


