package main

import ( "time"
)



func load_pred_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-load-pred"}
	s.init(nodes)

	// initialize specialized shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()

	load_pred_s := load_pred_shepherd{intlog_shepherd:intlog_s}
	load_pred_s.init()
	
	// start all management and coordination threads
	load_pred_s.deploy_musters()

	for _, l_m := range(load_pred_s.local_musters) {
		go load_pred_s.process_logs(l_m.id)
		go load_pred_s.start_exp(l_m.id)
	}
	go load_pred_s.process_control()

	time.Sleep(exp_timeout)
	for _, l_m := range(load_pred_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}




