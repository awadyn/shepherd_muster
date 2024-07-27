package main

import (
	"time"
)

/*********************************************/

func flink_main(n_ip string, n_cores uint8, pulse_server_port int, log_server_port int, ctrl_server_port int, coordinate_server_port int) {
	m := muster{}
	m.init(n_ip, n_cores)
	r_m := remote_muster{muster: m}
	r_m.init(pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port)

	flink_m := flink_muster{remote_muster: r_m}
	flink_m.init()
	flink_m.show()
	go flink_m.start_pulser()
	go flink_m.start_logger()
	go flink_m.start_controller()
	go flink_m.sync_new_ctrl()
	go flink_m.start_coordinator()
	<- flink_m.hb_chan
	for sheep_id, _ := range(flink_m.pasture) {
		for log_id, _ := range(flink_m.pasture[sheep_id].logs) { 
			go flink_m.log_manage(sheep_id, log_id) 
	//		go flink_m.ctrl_manage(sheep_id) 
		}
	}
	// cleanup
	time.Sleep(exp_timeout)
	flink_m.cleanup()
}



