#!/bin/bash

qps1=$1
update1=$2
qps2=$3
update2=$4

sudo killall mutilate # on  10.10.1.3
ssh 10.10.1.4 "sudo killall mutilate"
ssh 10.10.1.5 "sudo killall mutilate"
ssh 10.10.1.6 "sudo killall mutilate"
ssh 10.10.1.7 "sudo killall mutilate"
ssh 10.10.1.8 "sudo killall mutilate"
ssh 10.10.1.9 "sudo killall mutilate"

ssh 10.10.1.1 "sudo killall memcached"
#ssh 10.10.1.3 "sudo killall memcached"

sleep 1 

echo "START MEMCACHED"
ssh 10.10.1.1 "numactl --cpunodebind=0 --membind=0 taskset -c 0-8 ~/memcached/memcached -p 11211 -u nobody -t 8 -m 32768 -c 8192 -b 8192 -l 10.10.1.1 -B binary > mcd_11211.log 2>&1 < /dev/null &"
ssh 10.10.1.1 "numactl --cpunodebind=1 --membind=1 taskset -c 10-18 ~/memcached/memcached -p 11212 -u nobody -t 8 -m 32768 -c 8192 -b 8192 -l 10.10.1.1 -B binary > mcd_11212.log 2>&1 < /dev/null &"

echo "START LOAD GENERATION AGENTS"
ssh 10.10.1.4 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.5 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.6 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.7 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.8 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh 10.10.1.9 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
sleep 1

echo "LOAD MEMCACHED DATABASE"
taskset -c 0 ~/mutilate/mutilate -vv --binary -s 10.10.1.1:11211 --loadonly -K fb_key -V fb_value
taskset -c 0 ~/mutilate/mutilate -vv --binary -s 10.10.1.1:11212 --loadonly -K fb_key -V fb_value

mkdir -p exp/
rm -rf exp/*
echo > exp/mcd_11211.out
echo > exp/mcd_11212.out

sleep 1

echo "FLUSH IXGBE LOGS"
ssh 10.10.1.1 "mkdir -p ixgbe_logs/; cat /proc/ixgbe_stats/core/* | wc -l"

#echo "START MUSTHERD"
#ssh 10.10.1.1 "cd ~/shepherd_muster/shep_remote_muster/; taskset -c 0,10 /usr/local/go/bin/go run ./remote_muster/ 10.10.1.1 10.10.1.3 20 50051 50061 50071 50081 > muster.log 2>&1 < /dev/null &"
#sleep 3 && cd ~/shepherd_muster/shep_remote_muster/ && /usr/local/go/bin/go run ./shepherd/ >> shepherd.log &
#sleep 3 && cd ~/;
 
ssh 10.10.1.1 "cat /proc/ixgbe_stats/core/* | wc -l"

echo "START EXPERIMENT"
taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.1:11211 --agent={10.10.1.4,10.10.1.5,10.10.1.6} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=$update1 --qps=$qps1 --depth=128 --measure_connections=48 --measure_qps=2000 --time=30 >> exp/mcd_11211.out &
taskset -c 10 ~/mutilate/mutilate --binary -s 10.10.1.1:11212 --agent={10.10.1.7,10.10.1.8,10.10.1.9} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=$update2 --qps=$qps2 --depth=128 --measure_connections=48 --measure_qps=2000 --time=30 >> exp/mcd_11212.out &

# sleep and wait..
sleep 31
		
#ssh 10.10.1.1 "sudo killall remote_muster"
#sudo killall shepherd
#sleep 1
#cp -v ~/shepherd_muster/shep_remote_muster/mustherd-logs-muster-10.10.1.1/* exp/
#for i in {0..19}; do mv exp/ixgbe-log-core-$i-10.10.1.1 exp/core$i; done

ssh 10.10.1.1 'for i in {0..19}; do cat /proc/ixgbe_stats/core/$i > ixgbe_logs/core$i; done'
scp -r 10.10.1.1:~/ixgbe_logs/* exp/

ssh 10.10.1.4 "sudo killall mutilate"
ssh 10.10.1.5 "sudo killall mutilate"
ssh 10.10.1.6 "sudo killall mutilate"
ssh 10.10.1.7 "sudo killall mutilate"
ssh 10.10.1.8 "sudo killall mutilate"
ssh 10.10.1.9 "sudo killall mutilate"

ssh 10.10.1.1 "sudo killall memcached"
#ssh 10.10.1.3 "sudo killall memcached"


