#!/bin/bash

cd ~/
sudo apt install -y libevent-dev
echo "Downloading latest memcached.."
sleep 1
wget http://memcached.org/latest	
if ! [ -d "memcached_latest" ]; then mkdir memcached_latest; fi
tar -xf latest -C memcached_latest --strip-components=1
latest_version=$(cat memcached_latest/version.m4 | cut -d "," -f2 | grep -oP "\d+.\d+.\d+")
echo "Checking memcached.."
sleep 1
if [ -d "memcached" ]; then
	echo "Memcached found.."
	sleep 1
	current_version=$(cat ~/memcached/version.m4 | cut -d "," -f2 | grep -oP "\d+.\d+.\d+")
	if ! [[ $current_version == $latest_version  ]]; then
		echo "Memcached is not at latest version.. building latest.."
		sleep 1
		cd memcached_latest		
		./configure && make && make test && sudo make install
		cd ~/; mv ~/memcached_latest ~/memcached
	else
		echo "Memcached is at latest version.."
		sleep 1
		rm -rf ~/memcached_latest
		if ! [ -x ~/memcached/memcached ]; then
			echo "Building memcached.."
			cd memcached
			./configure && make && make test && sudo make install
		fi
	fi
else 
	echo "No memcached found.. building latest.."
	sleep 1
	cd memcached_latest
	./configure && make && make test && sudo make install
	cd ~/; mv ~/memcached_latest ~/memcached
fi

#sudo rmmod mlx4_ib;	
#sudo rmmod mlx4_core'

