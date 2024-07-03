mcd_server=$1
mutilate_server=$2
mutilate_agents=$3
myself="awadyn"


IFS=$IFS,
agents=()
for agent in $mutilate_agents; do agents+=($agent); done

echo "Memcached server: $mcd_server"
echo "Mutilate server: $mutilate_server"
echo "Mutilate agents: "
for agent in ${agents[@]}; do echo $agent; done

for dvfs in '0xc00' '0xe00' '0x1100' '0x1300' '0x1500' '0x1700' '0x1900' '0x1a00'; do

echo "Running memcached server on $mcd_server";
ssh $myself@$mcd_server "bash --login -c 'sudo pkill memcached; sudo pgrep -f go | xargs kill; sudo rmmod ixgbe'";
ssh $myself@$mcd_server "echo off | sudo tee /sys/devices/system/cpu/smt/control;
	echo "1" | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo;
	sudo wrmsr -a 0x199 $dvfs;
	sudo rdmsr -a 0x199;
	sudo insmod ~/linux-5.15.89/drivers/net/ethernet/intel/ixgbe/ixgbe.ko;	
	sudo ip link set dev enp3s0f0 up;
	sudo ip addr add 10.10.1.2/24 dev enp3s0f0;
	sleep 1;
	ip addr;
	taskset -c 0-15 ~/memcached_latest/memcached -u nobody -t 16 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_dump 2>&1 < /dev/null &";


echo "Running mutilate agents on $mutilate_agents";
ssh $myself@$mutilate_server "bash --login -c 'sudo pkill mutilate; sudo pgrep -f go | xargs kill'";
ssh $myself@$mutilate_server "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value";
for agent in ${agents[@]}; do
ssh $myself@$agent "sudo pkill mutilate;
	sleep 1;
	~/mutilate/mutilate --agentmode --threads=16 > mutilate_dump 2>&1 < /dev/null &";
done;


for qps in 50000 100000 200000 400000 600000; do
echo "QPS $qps..";

echo "Starting remote muster on $mcd_server..";
ssh $myself@$mcd_server	"bash --login -c 'cd ~/shepherd_muster/shep_remote_muster;
	go run remote_muster/* 10.10.1.2 10.10.1.1 16 50051 50061 50071 50081 > remote_muster_dump 2>&1 < /dev/null &'";

echo "Starting shepherd on $mutilate_server..";
ssh $myself@$mutilate_server "bash --login -c 'cd ~/shepherd_muster/shep_remote_muster;
	go run shepherd/* > shepherd_dump 2>&1 < /dev/null &'"; 

echo "Running mutilate bench with QPS $qps.."
ssh $myself@$mutilate_server "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4} --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=32 --measure_qps=2000 --qps=$qps --time=30 > mutilate_latency; 
	cat mutilate_latency;
	sleep 2";

echo "Done.. copying logs.."
ssh $myself@$mcd_server "bash --login -c 'sudo pgrep -f go | xargs kill'";
ssh $myself@$mutilate_server "bash --login -c 'sudo pgrep -f go | xargs kill'";
sleep 1;

scp -r $myself@$mutilate_server:~/shepherd_muster/logs/* mcd_ixgbe_muster_logs/$qps/;
scp -r $myself@$mutilate_server:~/mutilate_latency mcd_ixgbe_muster_logs/$qps/mutilate_latency_"$dvfs";
for i in {0..15}; do
mv -v mcd_ixgbe_muster_logs/$qps/muster-10.10.1.2_"$i"_1_0xc00.intlog mcd_ixgbe_muster_logs/$qps/"$dvfs"_"$i".intlog;
done;

done;

# cleanup after every dvfs setting
ssh $myself@$mcd_server 'sudo rmmod ixgbe; sudo pkill memcached'
ssh $myself@$mutilate_server 'sudo pkill mutilate'
for agent in ${agents[@]}; do
ssh $myself@$agent 'sudo pkill mutilate';
done;

done


