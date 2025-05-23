package main

import (
	"fmt"
	"os"
	"strconv"
)

/*********************************************/

var debug bool = false
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

	target_resources := make([]resource, 0)
	var i uint8
	for i = 0; i < ncores; i++ {
		if i < 9 { target_resources = append(target_resources, resource{label: "core", index: i+1}) }
		if i >= 9 { target_resources = append(target_resources, resource{label: "core", index: i+2}) }
	}
//	target_resources = append(target_resources, resource{label: "node", index: 0})

	remote_node := node{ip: n_ip, ip_idx: ip_idx, ncores: ncores, 
			    pulse_port: pulse_port, log_port: log_port, 
			    ctrl_port: ctrl_port, coordinate_port: coordinate_port,
		    	    resources: target_resources}


//	bayopt_intlog_main(remote_node)
	intlog_main(remote_node)
}



