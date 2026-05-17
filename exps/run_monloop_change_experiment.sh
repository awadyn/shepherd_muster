#!/bin/bash

echo "START MEMCACHED"
# start one mcd proc on node1
ssh 10.10.1.2 "taskset -c 0-15 ~/memcached/memcached -p 11211 -u nobody -t 16 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_11211.log 2>&1 < /dev/null &"

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

# start load generation for 1 mcd process on node1
taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2:11211 --noload --agent=10.10.1.3 --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.50 --qps=100000 --depth=128 --measure_connections=64 --measure_qps=2000 --time=30 >> exp/mcd_11211.out &

sleep 33

# read rapl
ssh 10.10.1.2 "./read_rapl_c6220.sh"


echo "SAVE IXGBE LOGS"
ssh 10.10.1.2 'for i in {0..15}; do cat /proc/ixgbe_stats/core/$i > ixgbe_logs/core$i; done'
scp -r 10.10.1.2:~/ixgbe_logs/* exp/
scp -r 10.10.1.2:~/rapl_* exp/

echo "KILLING SERVER AND AGENTS"
ssh 10.10.1.2 "sudo killall memcached"


