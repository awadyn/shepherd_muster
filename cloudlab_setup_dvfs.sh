node=$1
myself="awadyn"

ssh $myself@$node 'sudo apt update && sudo apt upgrade -y;
	sudo apt install msr-tools;
	sudo modprobe msr;
	echo off | sudo tee /sys/devices/system/cpu/smt/control;
	echo "1" | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo'

# NOTE: if modprobe fails, then probably msr is not set in .config 

ssh $myself@$node 'N=$(nproc);
	for i in $( seq 0 $N); do 
		if [ $i == $N ]; then break; fi; 
		echo "userspace" | sudo tee /sys/devices/system/cpu/cpu$i/cpufreq/scaling_governor; 
	done'


