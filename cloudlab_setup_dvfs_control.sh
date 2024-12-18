#!/bin/bash

echo "Downloading and probing msr-tools."
sleep 1
sudo apt install msr-tools
sudo modprobe msr

echo "Setting up dvfs governor to 'userspace governor'."
sleep 1
N=$(nproc)
for i in $( seq 0 $N); do 
	if [ $i == $N ]; then break; fi; 
	echo "userspace" | sudo tee /sys/devices/system/cpu/cpu$i/cpufreq/scaling_governor; 
done


