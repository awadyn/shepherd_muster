#!/bin/bash

port=11211
server=10.10.1.2
myip=10.10.1.1
server2=128.110.96.27
myip2=128.110.96.6
agent1=10.10.1.3
agent2=10.10.1.4
agent3=10.10.1.5

echo "START MEMCACHED 16 THREADS"
ssh $server "taskset -c 1-7,9-15 ~/memcached/memcached -p $port -u nobody -t 14 -m 32G -c 8192 -b 8192 -l $server -B binary > mcd_11211.log 2>&1 < /dev/null &"

echo "START LOAD GENERATION AGENTS"
ssh $agent1 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh $agent2 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"
ssh $agent3 "~/mutilate/mutilate --agentmode --threads=16 > agent.log 2>&1 < /dev/null &"

sleep 1
echo "LOAD MEMCACHED DATABASE"
taskset -c 0 ~/mutilate/mutilate -vv --binary -s $server:$port --loadonly -K fb_key -V fb_value

echo "CREATE TEMP EXP DIR"
mkdir -p exp/
rm -rf exp/*

sleep 1

update=$1
qps=$2

echo "START RUN UPDATE $update QPS $qps"
mkdir -p exp/$update\_$qps/

echo "START MUSTHERD"
ssh $server "cd ~/shepherd_muster/shep_remote_muster/; taskset -c 0,8 /usr/local/go/bin/go run remote_muster/* $server $myip 16 50051 50061 50071 50081 > muster.log 2>&1 < /dev/null &"
sleep 3 && cd ~/shepherd_muster/shep_remote_muster/;
/usr/local/go/bin/go run shepherd/* >> shepherd.log &
sleep 3 && cd ~/;

echo "START LOAD GENERATION"
# start load generation for 1 mcd process
taskset -c 0 ~/mutilate/mutilate --binary -s $server:$port --noload --agent={$agent1,$agent2,$agent3} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=$update --qps=$qps --depth=128 --measure_connections=48 --measure_qps=2000 --time=60 >> exp/$update\_$qps/leader.log &

sleep 65

echo "KILLING MUSTHERD"
ssh 10.10.1.2 "sudo killall go; sudo killall read_ixgbe_stats; sudo killall bayopt_main"
sudo killall go; sudo killall defs
sleep 3

echo "COPYING LOGS"
for c in {0..15}; do cp -v ~/shepherd_muster/shep_remote_muster/mustherd-logs-muster-$server/ixgbe-log-core-$c-$server exp/$update\_$qps/core$c; done
scp -r $agent1:~/agent.log exp/$update\_$qps/agent1.log
scp -r $agent2:~/agent.log exp/$update\_$qps/agent2.log
scp -r $agent3:~/agent.log exp/$update\_$qps/agent3.log

ssh $agent1 "sudo killall mutilate"
ssh $agent2 "sudo killall mutilate"
ssh $agent3 "sudo killall mutilate"
ssh $server "sudo killall memcached"




