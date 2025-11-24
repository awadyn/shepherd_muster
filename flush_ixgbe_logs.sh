#!/bin/bash

for i in {0..15}; do 
	cat /proc/ixgbe_stats/core/$i | wc -l; 
done
