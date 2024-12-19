# MustHerd
MustHerd is a distributed framework that enables learning-based control decision making on one or more target nodes.
MustHerd's main components are a centralized shepherd and distributed musters.

(See description in  https://github.com/awadyn/shepherd_muster.wiki.git)

## Preparing MustHerd Environment
#### Assuming one shepherd node and at least one muster node:
```bash
ssh user@node-shepherd 'sudo apt update && sudo apt upgrade -y'
ssh user@node-muster-x 'sudo apt update && sudo apt upgrade -y'
```

#### Clone MustHerd code base on shepherd node and all muster nodes:
```bash
user@node:$ git clone git@github.com:awadyn/shepherd_muster.git
user@node:$ cd shepherd_muster; ./cloudlab_setup_golang.sh
```

Running above script checks for a compatible golang version:
```bash
user@node:$ go_version=1.22.4							// compatible golang version
user@node:$ which go								// checks if go runtime is installed
user@node:$ go version 								// checks go version
user@node:$ sudo rm -rf /usr/local/bin/go 					// remove current go version
user@node:$ wget https://go.dev/dl/go$go_version.linux-amd64.tar.gz		// download go version
user@node:$ sudo tar -C /usr/local -xzf go$go_version.linux-amd64.tar.gz	// install go locally
user@node:$ echo 'export PATH=$PATH:/usr/local/go/bin' >> .bashrc		// add go binary to bash shell environment
```
After running the above script, exit and re-enter shell session to apply bashrc changes.


## Preparing Example Native Logger Environment
#### First build and install compatible kernel version:
```bash
user@node:$ cd shepherd_muster; ./cloudlab_setup_ixgbe_kernel.sh
```
Running above script downloads, installs, and builds kernel version:
```bash
user@node:$ kernel=$(uname -r)								// reads kernel version
user@node:$ wget https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.15.89.tar.xz	// downloads compatible linux kernel
user@node:$ tar -xf linux-5.15.89.tar.xz
user@node:$ cd linux-5.15.89
user@node:$ cp -v /boot/config-$(uname -r) .config 					// copies current kernel config to compatible kernel code base
user@node:$ make localmodconfig								// building kernel..
user@node:$ scripts/config --disable SYSTEM_TRUSTED_KEYS			
user@node:$ scripts/config --disable SYSTEM_REVOCATION_KEYS
user@node:$ scripts/config --set-str CONFIG_SYSTEM_TRUSTED_KEYS ""
user@node:$ scripts/config --set-str CONFIG_SYSTEM_REVOCATION_KEYS ""
user@node:$ fakeroot make -j8
user@node:$ sudo make modules_install
user@node:$ sudo make install
user@node:$ sudo reboot									// see new kernel version after reboot

```
#### Then download and build modified ixgbe driver into compatible kernel:
```bash
user@node:$ cd shepherd_muster; ./cloudlab_setup_ixgbe_logger.sh
```
Running the above script rebuilds kernel with modified ixgbe driver:
```bash
user@node:$ git clone https://github.com/handong32/intlog.git
user@node:$ cp -r ~/intlog/linux/linux-5.15.89/drivers/net/ ~/linux-5.15.89/drivers/
user@node:$ cd linux-5.15.89
user@node:$ fakeroot make -j8
```

Then, newly built ixgbe driver is loaded:
```bash
user@node:$ sudo rmmod ixgbe
user@node:$ sudo insmod ~/linux-5.15.89/drivers/net/ethernet/intel/ixgbe/ixgbe.ko
user@node:$ ieth=$(sudo dmesg | grep "ixgbe" | grep "renamed from eth0" | tail -n 2 | head -n 1 | grep -oP "enp\ds\df\d")
user@node:$ num=$(uname -a | grep -oP "node\d" | grep -oP "\d")
user@node:$ node=$(($num + 1))
user@node:$ ip="10.10.1.$node"
user@node:$ sudo ip link set dev $ieth up
user@node:$ sudo ip addr add $ip/24 dev $ieth
```

It also sets system hardware settings that can jeopardize correct behavior of ixgbe driver:
```bash
user@node:$ echo off | sudo tee /sys/devices/system/cpu/smt/control
user@node:$ echo "1" | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo
user@node:$ sudo killall irqbalance
user@node:$ sudo ~/shepherd_muster/intel_set_irq_affinity.sh <ieth>
```

Finally, it checks ixgbe driver stats:
```bash
user@node:$ sudo ~/shepherd_muster/intel_set_irq_affinity.sh $ieth
user@node:$ for i in {0..15}; do cat /proc/ixgbe_stats/core/$i; echo; done
```

## Preparing Example Control Environment
#### Here we enable userspace dynamic voltage/frequency scaling on a muster node:
```bash
user@muster:$ cd shepherd_muster/; ./cloudlab_setup_dvfs_control.sh
```

Running the above script re-configures kernel to enable x86 msr manipulation (if disabled):
```bash
user@node:$ cd ~/linux-5.15.89
user@node:$ scripts/config --enable CONFIG_X86_MSR
user@node:$ fakeroot make -j8
user@node:$ sudo apt install msr-tools
user@node:$ sudo modprobe msr
```

It then sets userspace scaling governor for all active cores:
```bash
user@node:$ N=$(nproc)
user@node:$ for i in $( seq 0 $N); do if [ $i == $N ]; then break; fi; echo "userspace" | sudo tee /sys/devices/system/cpu/cpu$i/cpufreq/scaling_governor; done
```

## Preparing Example Workload Environment
#### Here we enable memcached as a workload on a muster node:
```bash
user@muster:$ cd shepherd_muster/; ./cloudlab_setup_memcached.sh
user@muster:$ cd ~/; ./memcached/memcached --help
```

#### Here we enable mutilate load generator to drive memcached workload:
This must be done on all nodes involved in a mutilate cluster load generation:
```bash
user@node:$ cd shepherd_muster/; ./cloudlab_setup_mutilate.sh
user@node:$ cd ~/; ./mutilate/mutilate --help
```

## Preparing Example Optimizer Environment
This must be done on the shepherd node or the node responsible for running optimization processes if different from the shepherd node.
This particular example prepares a python environment for a bayesian optimizer service:
```bash
user@node:$ cd shepherd_muster/shep_remote_muster/
user@node:$ sudo apt install python3-virtualenv
user@node:$ virtualenv mustherd-venv
user@node:$ source mustherd-venv/bin/activate
(mustherd-venv) user@node:$ pip3 install grpcio
(mustherd-venv) user@node:$ pip3 install google-api-python-client
(mustherd-venv) user@node:$ pip3 install ax-platform
```

## Running MustHerd Memcached Test
#### Checking shepherd-to-muster connections and pulsing:
First, start optimizer service on the shepherd node, terminal A:
```bash
user@shepherd:$ cd shepherd_muster/shep_remote_muster
user@node:$ source mustherd-venv/bin/activate
(mustherd-venv) user@node:$ python3 optimizer_server.py
```

Second, start shepherd on the shepherd node, terminal B:
```bash
user@shepherd:$ cd shepherd_muster/shep_remote_muster
user@shepherd:$ go run shepherd/*
```

First, start Memcached server on the muster nodes, terminal A:
```bash
user@node:$ taskset -c 0-15 ~/memcached_latest/memcached -u nobody -t 16 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary
```

Second, start remote muster on the muster nodes, terminal B:
```bash
user@muster:$ cd shepherd_muster/shep_remote_muster
user@muster:$ #go run remote_muster/* <muster_ip> <shepherd_ip> <num_cores> <pluse_port> <log_port> <ctrl_port> <coord_port> <optional_ip_idx>
user@muster:$ go run remote_muster/* 10.10.1.2 10.10.1.1 16 50051 50061 50071 50081
```

(Other nodes in this experiment may be used as mutilate load generation agents. MustHerd will not run on these nodes.)
