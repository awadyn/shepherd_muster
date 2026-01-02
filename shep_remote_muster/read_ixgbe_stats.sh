#!/bin/bash

src=$1
dest=$2

cat $src > /dev/null
while true
do 
#	cat $src >> $dest
#	cat $src  | cut -d ' ' -f 3,17 >> $dest			# rx_bytes,timestamp
	cat $src  | cut -d ' ' -f 3,6,7,8,9,17 >> $dest		# rx_bytes,instructions,cycles,ref_cycles,llc_miss,timestamp
	sleep 1
done
