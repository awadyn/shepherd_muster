node=$1
myself="awadyn"

#ssh $myself@$node 'sudo apt update && sudo apt upgrade -y'

kernel=$(ssh $myself@$node 'uname -r')
if [[ $kernel != "5.15.89" ]]; 
then 
echo "Kernel version $kernel is bad for building intlogger. Installing 5.15.89 instead.";
ssh $myself@$node 'sudo apt install -y fakeroot dwarves flex bison libssl-dev libelf-dev'
ssh $myself@$node 'wget https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.15.89.tar.xz; 
	tar -xf linux-5.15.89.tar.xz; 					\
	cd linux-5.15.89; 						\
	cp -v /boot/config-$(uname -r) .config;  			\
	make localmodconfig;						\
	scripts/config --disable SYSTEM_TRUSTED_KEYS; 			\
	scripts/config --disable SYSTEM_REVOCATION_KEYS; 		\
	scripts/config --set-str CONFIG_SYSTEM_TRUSTED_KEYS ""; 	\
	scripts/config --set-str CONFIG_SYSTEM_REVOCATION_KEYS ""; 	\
	fakeroot make -j8;						\
	sudo make modules_install;					\
	sudo make install;						\
	sudo reboot'
sleep 30
fi


echo "Kernel version $kernel found. Installing intlogger.."
ssh $myself@$node 'sudo rmmod ixgbe;							\
	if [ -d intlog ]; then echo "intlog source exists.. skipping clone.."; 		\
	else git clone https://github.com/handong32/intlog.git;	fi;			\
	linux_dir="linux-5.15.89/drivers/net/ethernet/intel/ixgbe";			\
	intlog_dir="intlog/linux/"$linux_dir;						\
	delta=$(cmp $linux_dir/ixgbe_main.c $intlog_dir/ixgbe_main.c);			\
	if [ $delta ]; then 								\
	echo "linux ixgbe and intlog ixgbe differ"; 					\
	cp -r ~/intlog/linux/linux-5.15.89/drivers/net/ ~/linux-5.15.89/drivers/; 	\
	cd linux-5.15.89; 								\
	fakeroot make -j8; fi';

echo "Disabling hyperthreads and turboboost.."
ssh $myself@$node 'echo off | sudo tee /sys/devices/system/cpu/smt/control;
	echo "1" | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo'

echo "Starting intlogger.."
ssh $myself@$node 'sudo insmod ~/linux-5.15.89/drivers/net/ethernet/intel/ixgbe/ixgbe.ko;		
	ieth=$(sudo dmesg | grep "ixgbe" | grep "renamed from eth0" | tail -n 2 | head -n 1 | grep -oP "enp\ds\df\d");
	num=$(uname -a | grep -oP "node\d" | grep -oP "\d");
	node=$(($num + 1));
	ip="10.10.1.$node";	
	echo "intlogger network interface: $ieth";							
	echo "intlogger network ip: $ip";								
	sudo ip link set dev $ieth up;						
	sudo ip addr add $ip dev $ieth'

echo "Disabling irqbalance and setting irq affinity.."
ssh $myself@$node 'sudo killall irqbalance;
	ieth=$(sudo dmesg | grep "ixgbe" | grep "renamed from eth0" | tail -n 2 | head -n 1 | grep -oP "enp\ds\df\d");
	sudo ./intel_set_irq_affinity.sh $ieth'

echo "Testing intlogger.."
ssh $myself@$node 'for i in {0..15}; do cat /proc/ixgbe_stats/core/$i; echo; done'


