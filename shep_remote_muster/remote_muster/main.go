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

	bayopt_m := bayopt_muster{remote_muster: r_m}
	bayopt_m.init()
	bayopt_m.show()
	go bayopt_m.start_pulser()
	go bayopt_m.start_logger()
	go bayopt_m.start_controller()
	go bayopt_m.sync_new_ctrl()
	go bayopt_m.start_coordinator()
	<- bayopt_m.hb_chan
	for sheep_id, _ := range(bayopt_m.pasture) {
		for log_id, _ := range(bayopt_m.pasture[sheep_id].logs) { 
			go bayopt_m.log_manage(sheep_id, log_id) 
			go bayopt_m.ctrl_manage(sheep_id) 
		}
	}
	// cleanup
	time.Sleep(exp_timeout)
	bayopt_m.cleanup()

//	intlog_m := intlog_muster{remote_muster: r_m}
//	intlog_m.init()
//	intlog_m.start_native_logger()
//	go intlog_m.start_pulser()
//	go intlog_m.start_logger()
//	//go intlog_m.start_controller()
//	//go intlog_m.handle_new_ctrl()
//	// cleanup
//	time.Sleep(exp_timeout)
//	intlog_m.cleanup()
}



