package main

import (
	"fmt"
	"os"
	"strconv"
)

/*********************************************/

var debug bool = true
//var debug bool = false
var mirror_ip string

func main() {
	n_ip := os.Args[1]
	mirror_ip = os.Args[2]
	arg3, err := strconv.Atoi(os.Args[3])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_cores argument: %v\n", err)}
	ncores := uint8(arg3)
	pulse_port, err := strconv.Atoi(os.Args[4])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	log_port, err := strconv.Atoi(os.Args[5])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	ctrl_port, err := strconv.Atoi(os.Args[6])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	coordinate_port, err := strconv.Atoi(os.Args[7])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	var ip_idx int = -1
	if len(os.Args) > 8 { 
		ip_idx, err = strconv.Atoi(os.Args[8])
		if err != nil {fmt.Printf("** ** ** ERROR: bad ip_idx argument: %v\n", err)}
	}

	target_resources := setup_target_resources_c8220(ncores)
	remote_node := node{ip: n_ip, ip_idx: ip_idx, ncores: ncores, 
			    pulse_port: pulse_port, log_port: log_port, 
			    ctrl_port: ctrl_port, coordinate_port: coordinate_port,
		    	    resources: target_resources}

	stats_main(remote_node)
}



