qps=$1
for itrd in 1 2 10 50 100 150 200 250 300 350 400; do
	for dvfs in 0xc00 0xe00 0x1100 0x1300 0x1500 0x1700 0x1900 0x1a00 0x1c00; do 
		ssh 10.10.1.2 "taskset -c 0-8,10-18 ~/memcached/memcached -u nobody -t 18 -m 32G -c 8192 -b 8192 -l 10.10.1.2 -B binary > mcd_dump 2>&1 < /dev/null &"
		sleep 1;
		taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value;
		for i in 3 4 5 6; do 
			ssh 10.10.1.$i "~/mutilate/mutilate --agentmode --threads=16 > mutilate_dump 2>&1 < /dev/null &";
		done
		sleep 1;

		echo "------------- ITRD: $itrd -- DVFS: $dvfs";
		ssh 10.10.1.2 "sudo wrmsr -a 0x199 $dvfs; sudo ethtool -C enp130s0f0 rx-usecs $itrd; sleep 1; sudo ./read_rapl_start.sh"; 
		taskset -c 0-7,10-17 ~/mutilate/mutilate --binary -s 10.10.1.2 --noload --agent={10.10.1.3,10.10.1.4,10.10.1.5,10.10.1.6} --threads=16 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_connections=4 --measure_qps=2000 --qps=$qps --time=30;
		ssh 10.10.1.2 "sudo ./read_rapl.sh"; 
		
		ssh 10.10.1.2 "sudo killall memcached";
		for i in 3 4 5 6; do 
			ssh 10.10.1.$i "sudo killall mutilate";
		done

		
		echo; 
		scp -r 10.10.1.2:~/rapl_pkg.txt .; 
		./compute_rapl_joules.sh; 
		echo; echo "--------------------------------"; echo;
	        sleep 2;	
	done
done
