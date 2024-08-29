package main

import (
	"os"
	"time"
)

/*********************************************/

func flink_main(n_ip string, n_cores uint8, pulse_server_port int, log_server_port int, ctrl_server_port int, coordinate_server_port int, ip_idx string) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	m0 := muster{}
	m0.init(n_ip, n_cores, ip_idx)
	m0.id = "worker-" + m0.id
	r_m_0 := remote_muster{muster: m0}
	r_m_0.init(pulse_server_port, log_server_port, ctrl_server_port, coordinate_server_port)
//	r_m_0.show()
 
	m1 := muster{}
	m1.init(n_ip, n_cores, ip_idx)
	m1.id = "source-" + m1.id
	r_m_1 := remote_muster{muster: m1}
	r_m_1.init(pulse_server_port+1, log_server_port+1, ctrl_server_port+1, coordinate_server_port+1)
//	r_m_1.show() 

	logs_dir := home_dir + "/" + r_m_0.id + "_" + r_m_1.id + "_flink_logs/"

	bayopt_m := bayopt_muster{remote_muster: r_m_0, logs_dir: logs_dir}
	_, err = os.Stat(logs_dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(logs_dir, 0777)
		if err != nil { panic(err) }
	}

	worker_m := flink_worker_muster{bayopt_muster: bayopt_m}
	worker_m.init()
	worker_m.show()

	source_m := flink_source_muster{remote_muster: r_m_1, logs_dir: logs_dir}
	source_m.init()
	source_m.show()


	go worker_m.start_pulser()
	go source_m.start_pulser()
	go worker_m.start_controller()
	go worker_m.sync_new_ctrl()
	go worker_m.start_coordinator()
	go source_m.start_coordinator()
	worker_m.start_logger()
	source_m.start_logger()

	for sheep_id, _ := range(worker_m.pasture) {
		go worker_m.log_manage(sheep_id) 
		go worker_m.ctrl_manage(sheep_id) 
	}
	for sheep_id, _ := range(source_m.pasture) {
		go source_m.log_manage(sheep_id) 
	}

	// cleanup
	time.Sleep(exp_timeout)
//	flink_m.cleanup()
}



