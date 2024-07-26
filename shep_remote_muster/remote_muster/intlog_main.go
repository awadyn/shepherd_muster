package main

import (
	"time"
)

/*********************************************/

func intlog_main(n_ip string, n_cores uint8, pulse_server_port int, log_server_port int, ctrl_server_port int, coordinate_server_port int) {
	m := muster{}
	m.init(n_ip, n_cores)
	r_m := remote_muster{muster: m}
	r_m.init(pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port)

	intlog_m := intlog_muster{remote_muster: r_m}
	intlog_m.init()
	intlog_m.start_native_logger()
	go intlog_m.start_pulser()
	go intlog_m.start_logger()
	//go intlog_m.start_controller()
	//go intlog_m.handle_new_ctrl()
	// cleanup
	time.Sleep(exp_timeout)
	intlog_m.cleanup()
}



