#!/bin/bash

# set start itrd = 1
ssh 128.110.96.38 "sudo ethtool -C enp3s0f0 rx-usecs 1; ./flush_ixgbe_logs.sh &> /dev/null" 

# start mutilate load generation
taskset -c 0-7 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.5} --threads=8 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=1000000 --time=30 & 

# sleep x; change itrd; repeat y times
for itrd in 100 20 300 50 2 200; do
	#taskset -c 8 echo $itrd;
	taskset -c 8 ssh 128.110.96.38 "taskset -c 7 sudo ethtool -C enp3s0f0 rx-usecs $itrd";
       	taskset -c 8 sleep 5;
done


# scp ixgbe logs
ssh 128.110.96.38 "./cat_ixgbe_logs.sh"
scp -r 128.110.96.38:~/ixgbe_logs/ .


