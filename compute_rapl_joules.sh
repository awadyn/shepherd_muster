#!/bin/bash

rapl_file=$1

j0=$(awk 'NR % 2 == 0' $rapl_file | head -n 1)
j00=$((16#$j0))
j1=$(awk 'NR % 2 == 0' $rapl_file | tail -n 1)
j01=$((16#$j1)) 

if (( $j00 > $j01 )); then
	diff0=$(( 2**32 - 1 - $j00 + $j01 ));
else
	diff0=$(( $j01 - $j00 ));
fi


j0=$(awk 'NR % 2 == 1' $rapl_file | head -n 2 | tail -n 1)
j10=$((16#$j0))
j1=$(awk 'NR % 2 == 1' $rapl_file | tail -n 1)
j11=$((16#$j1))

if (( $j10 > $j11 )); then
	diff1=$(( 2**32 - 1 - $j10 + $j11 ));
else
	diff1=$(( $j11 - $j10 ));
fi

sum=$(( $diff0 + $diff1 ))


#rapl readings: 3580202344 3672345222  --  907652021 979153127
#pkg0: 92142878 -- pkg1: 71501106 -- sum: 163643984
echo "rapl readings:" $j00 $j01 " -- " $j10 $j11
echo "pkg0:" $diff0 "-- pkg1:" $diff1 "-- sum:" $sum



