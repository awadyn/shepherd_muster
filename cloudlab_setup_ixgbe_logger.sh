#!/bin/bash

cd ~/
kernel=$(uname -r)
echo "Kernel version $kernel found. Installing intlogger.."
sleep 2
if [ ! -d "~/intlog" ]; then
git clone https://github.com/handong32/intlog.git
cp -r ~/intlog/linux/linux-5.15.89/drivers/net/ ~/linux-5.15.89/drivers/
cd linux-5.15.89
fakeroot make -j8
fi

sleep 2
echo "Disabling hyperthreads and turboboost.."
echo off | sudo tee /sys/devices/system/cpu/smt/control
echo "1" | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo

sleep 2
echo "Loading ixgbe interrupt logger.."
sudo rmmod ixgbe
sudo insmod ~/linux-5.15.89/drivers/net/ethernet/intel/ixgbe/ixgbe.ko
ieth=$(sudo dmesg | grep "ixgbe" | grep "renamed from eth0" | tail -n 2 | head -n 1 | grep -oP "enp\ds\df\d")
num=$(uname -a | grep -oP "node\d" | grep -oP "\d")
node=$(($num + 1))
ip="10.10.1.$node"
echo "ixgbe interrupt logger interface: $ieth"
echo "ixgbe interrupt logger ip: $ip"
sudo ip link set dev $ieth up
sudo ip addr add $ip/24 dev $ieth

sleep 2
echo "Disabling irqbalance and setting irq affinity.."
sudo killall irqbalance
sudo ~/shepherd_muster/intel_set_irq_affinity.sh $ieth

sleep 2
echo "Testing intlogger.."
for i in {0..15}; do cat /proc/ixgbe_stats/core/$i; echo; done


