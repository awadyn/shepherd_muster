package main

import ( "time"
	 "os" 
)

/**************************************/


func nop_main(nodes []node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	// initialize generic shepherd
	s := shepherd{id: "sheperd-nop"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	nop_s := nop_shepherd{shepherd:s,
			      logs_dir: home_dir + "/shepherd_nop_logs/"}
	nop_s.init()
	nop_s.init_local()
	
	// start all management and coordination threads
	nop_s.deploy_musters()
	go nop_s.listen_heartbeats()

	time.Sleep(exp_timeout)

	for _, l_m := range(nop_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}





