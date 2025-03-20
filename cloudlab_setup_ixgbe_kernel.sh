#!/bin/bash

cd ~/
kernel=$(uname -r)
if [[ $kernel != "5.15.89" ]]; 
then 
echo "Kernel version $kernel is bad for building intlogger. Installing 5.15.89 instead."
sleep 2
sudo apt install -y fakeroot dwarves flex bison libssl-dev libelf-dev
wget https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.15.89.tar.xz
tar -xf linux-5.15.89.tar.xz
cd linux-5.15.89
cp -v /boot/config-$(uname -r) .config 
yes "" | make localmodconfig
scripts/config --enable CONFIG_X86_MSR
scripts/config --disable SYSTEM_TRUSTED_KEYS
scripts/config --disable SYSTEM_REVOCATION_KEYS
scripts/config --set-str CONFIG_SYSTEM_TRUSTED_KEYS ""
scripts/config --set-str CONFIG_SYSTEM_REVOCATION_KEYS ""
fakeroot make -j8
sudo make modules_install
sudo make install
sudo reboot
fi

