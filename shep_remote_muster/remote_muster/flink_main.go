package main

//import (
//	"time"
//)
//
///*********************************************/
//
//func flink_main(n_ip string, n_cores uint8, pulse_server_port int, log_server_port int, ctrl_server_port int, coordinate_server_port int) {
//	m := muster{}
//	m.init(n_ip, n_cores)
//	r_m := remote_muster{muster: m}
//	r_m.init(pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port)
//
//	flink_m := flink_muster{remote_muster: r_m}
//	flink_m.init()
//
//	energy_m := flink_energy_muster{flink_muster: flink_m}
//	energy_m.init()
//	energy_m.show()
//
//	go energy_m.start_pulser()
//	go energy_m.start_logger()
//	go energy_m.start_controller()
//	go energy_m.sync_new_ctrl()
//	go energy_m.start_coordinator()
//	<- energy_m.hb_chan
//	for sheep_id, _ := range(energy_m.pasture) {
//		for log_id, _ := range(energy_m.pasture[sheep_id].logs) { 
//			go energy_m.log_manage(sheep_id, log_id) 
//	//		go energy_m.ctrl_manage(sheep_id) 
//		}
//	}
//	// cleanup
//	time.Sleep(exp_timeout)
//	energy_m.cleanup()
//}



