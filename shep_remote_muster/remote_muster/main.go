package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/*********************************************/

func main() {
	n_ip := os.Args[1]
	mirror_ip = os.Args[2]
	n_cores, err := strconv.Atoi(os.Args[3])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_cores argument: %v\n", err)}
	pulse_server_port, err := strconv.Atoi(os.Args[4])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	log_server_port := os.Args[5]
	ctrl_server_port, err := strconv.Atoi(os.Args[6])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	coordinate_server_port, err := strconv.Atoi(os.Args[7])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}

	m := muster{}
	m.init(n_ip, n_cores)

	r_m := remote_muster{muster: m}
	r_m.init(n_ip, n_cores, pulse_server_port, ctrl_server_port, log_server_port, coordinate_server_port)
	r_m.show()

	bayopt_m := bayopt_muster{remote_muster: r_m}
	bayopt_m.init()

	bayopt_m.start_native_logger()
	go bayopt_m.start_coordinator()

	go bayopt_m.start_pulser()
	go bayopt_m.start_logger()
	go bayopt_m.start_controller()
	go bayopt_m.handle_new_ctrl()

	// cleanup
//	go test_m.wait_done()
//	<- test_m.exit_chan
	time.Sleep(exp_timeout)
	bayopt_m.cleanup()
}


