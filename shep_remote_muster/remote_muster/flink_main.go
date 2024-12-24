package main

import (
	"os"
	"time"
)

/*********************************************/

func flink_main(n node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	m0 := muster{node: n}
	m0.init()
	m0.init_remote()
	r_m_0 := remote_muster{muster: m0}
	r_m_0.init()
 
	m1 := muster{node: n}
	m1.init()
	m1.init_remote()
	r_m_1 := remote_muster{muster: m1}
	r_m_1.init()




	logs_dir := home_dir + "/" + r_m_0.id + "." + r_m_1.id + ".flink_logs/"
	bayopt_m := bayopt_muster{remote_muster: r_m_0}
	bayopt_m.logs_dir = logs_dir
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
	go worker_m.start_coordinator()
	go source_m.start_coordinator()
	worker_m.start_logger()
	source_m.start_logger()

	for sheep_id, _ := range(worker_m.pasture) {
		go worker_m.log_manage(sheep_id, logs_dir, ixgbe_native_log) 
		go worker_m.ctrl_manage(sheep_id) 
	}
	for sheep_id, _ := range(source_m.pasture) {
		go source_m.log_manage(sheep_id, logs_dir, ixgbe_native_log) 
	}

	// cleanup
	time.Sleep(exp_timeout)
	source_m.cleanup()
	worker_m.cleanup()
}



