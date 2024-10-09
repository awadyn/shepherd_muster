#!/bin/bash

src=$1
dest=$2

cat $src > /dev/null
while true
do 
	sleep 4
	cat $src >> $dest
done
