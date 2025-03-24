#!/bin/bash

sleep 1;
kernel=$(uname -r)
if [[ $kernel != "5.15.89" ]]; 
then 
sudo apt update && sudo apt upgrade -y;
./cloudlab_setup_ixgbe_kernel.sh;
fi

sleep 1;
./cloudlab_setup_golang.sh;
sleep 1;
./cloudlab_setup_memcached.sh;
sleep 1;
./cloudlab_setup_dvfs_control.sh;
sleep 1;
./cloudlab_setup_ixgbe_logger.sh;

