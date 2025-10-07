#!/bin/bash

# sleep x; change number of memcached servers; repeat
for n in $(seq 0 2); do
	p=$((11212 + $n));
	echo $n $p;
	ssh 128.110.96.38 "taskset -c 0-6,8-14 ~/memcached/memcached -p $p -u nobody -t 14 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_dump_$p 2>&1 < /dev/null &";
       	echo;
done

# set start itrd = 1
ssh 128.110.96.38 "sudo ethtool -C enp3s0f0 rx-usecs 1; ./flush_ixgbe_logs.sh &> /dev/null" 

# start mutilate load generation
taskset -c 0-7 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.5,10.10.1.6} --threads=8 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=600000 --time=30 

# scp ixgbe logs
ssh 128.110.96.38 "./cat_ixgbe_logs.sh"
scp -r 128.110.96.38:~/ixgbe_logs/ .


