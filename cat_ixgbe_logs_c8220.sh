#!/bin/bash

dir=$1

for i in {0..19}; do
	if [[ $i -eq 9 ]]; then continue; fi
	if [[ $i -eq 19 ]]; then continue; fi
	cat /proc/ixgbe_stats/core/$i | cut -d ' ' -f 3,17 > $dir/core_$i;
done
