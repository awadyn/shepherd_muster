#!/bin/bash

dir=$1

for i in {0..15}; do
	if [[ $i -eq 7 ]]; then continue; fi
	if [[ $i -eq 15 ]]; then continue; fi
	cat /proc/ixgbe_stats/core/$i | cut -d ' ' -f 3,17 > $dir/core_$i;
done
