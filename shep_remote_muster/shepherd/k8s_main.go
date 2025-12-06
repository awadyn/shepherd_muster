package main

import ( 
	"time"
)



func k8s_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-k8s"}
	s.init(nodes)

	// initialize specialized shepherd
	k8s_s := k8s_shepherd{shepherd: s}
	k8s_s.init()
	
	// start all management and coordination threads
	k8s_s.deploy_musters()

	for _, l_m := range(k8s_s.local_musters) {
		go k8s_s.process_logs(l_m.id)
	}
	go k8s_s.process_control()

	time.Sleep(exp_timeout)
	for _, l_m := range(k8s_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { f.Close() }
		}
	}
}




