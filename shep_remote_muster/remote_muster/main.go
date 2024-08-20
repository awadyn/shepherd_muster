package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/*********************************************/

var exp_timeout time.Duration = time.Second * 75
var mirror_ip string;

func main() {
	n_ip := os.Args[1]
	mirror_ip = os.Args[2]
	arg3, err := strconv.Atoi(os.Args[3])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_cores argument: %v\n", err)}
	n_cores := uint8(arg3)

	pulse_server_port, err := strconv.Atoi(os.Args[4])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	log_server_port, err := strconv.Atoi(os.Args[5])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	ctrl_server_port, err := strconv.Atoi(os.Args[6])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	coordinate_server_port, err := strconv.Atoi(os.Args[7])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	var ip_idx string = ""
	if len(os.Args) > 8 { ip_idx = os.Args[8] }

//	flink_main(n_ip, n_cores, pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port)
	bayopt_main(n_ip, n_cores, pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port, ip_idx)
}



