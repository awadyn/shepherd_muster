package main

import ( 
//	"fmt"
	"time"
//	"math"
//	"slices"
//	"sort"
//	"google.golang.org/protobuf/types/known/anypb"
//	"google.golang.org/protobuf/types/known/wrapperspb"
)




func load_pred_bayopt_main(nodes []node) {
	// initialize generic shepherd
	s := shepherd{id: "sheperd-load-pred"}
	s.init(nodes)
	
	// initialize specialized shepherd
	intlog_s := intlog_shepherd{shepherd:s}
	intlog_s.init()
	load_pred_bayopt_s := load_pred_bayopt_shepherd{intlog_shepherd:intlog_s}
	load_pred_bayopt_s.init()

	load_pred_bayopt_s.show()
	
	// start all management and coordination threads
	load_pred_bayopt_s.deploy_musters()
	for _, l_m := range(load_pred_bayopt_s.local_musters) {
		go load_pred_bayopt_s.process_logs(l_m.id)
		go load_pred_bayopt_s.process_control(l_m.id)
	}

	// start logging and optimization threads for this particular experiment
	go load_pred_bayopt_s.start_exp()

	time.Sleep(exp_timeout)
	load_pred_bayopt_s.cleanup()
}
