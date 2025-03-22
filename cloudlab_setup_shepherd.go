#!/bin/bash

sudo apt update && sudo apt upgrade -y;
sleep 1;
./cloudlab_setup_golang.sh;
sleep 1;
./cloudlab_setup_mutilate.sh


