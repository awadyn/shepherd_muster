package main

import (
	"time"
)

/*********************************************/

func bayopt_main(n_ip string, n_cores uint8, pulse_server_port int, log_server_port int, ctrl_server_port int, coordinate_server_port int) {
	m := muster{}
	m.init(n_ip, n_cores)
	r_m := remote_muster{muster: m}
	r_m.init(pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port)

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
}



