package main

import ( "time"
)

/**************************************/


func intlog_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-intlog"}
	s.init(nodes)

	// initialize specialized shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()
	
	// start all management and coordination threads
	intlog_s.deploy_musters()

	for _, l_m := range(intlog_s.local_musters) {
		go intlog_s.process_logs(l_m.id)
	}

	go intlog_s.start_exp()

	time.Sleep(exp_timeout)
	intlog_s.cleanup()
}





