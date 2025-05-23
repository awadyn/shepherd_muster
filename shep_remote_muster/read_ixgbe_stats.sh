#!/bin/bash

src=$1
dest=$2

cat $src > /dev/null
while true
do 
	cat $src >> $dest
	sleep 1
done
