package main

import ( 
	"time"
	"strconv"
	"os"
	"os/exec" 
//	"fmt"
)

/**************************************/

func (bayopt_s *bayopt_shepherd) run_workload(m_id string) {
	l_m := bayopt_s.local_musters[m_id]
	<- l_m.hb_chan

//	// establish connection with remote muster and
//	// get current remote node control state
//	for _, sheep := range(l_m.pasture) {
//		for _, ctrl := range(sheep.controls) {
//			l_m.request_ctrl_chan <- []string{sheep.id, ctrl.id}
//		}
//	}
	// at this point, ctrl values are set in local muster representation
	bayopt_s.init_log_files(bayopt_s.logs_dir)

	for iter := 0; iter < 2; iter ++ {
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "start"}
			}
		}

		cmd := exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s 10.10.1.2 --loadonly -K fb_key -V fb_value")
		if err := cmd.Run(); err != nil { panic(err) }
		cmd = exec.Command("bash", "-c", "taskset -c 0 ~/mutilate/mutilate --binary -s " + l_m.ip + " --noload --threads=1 --keysize=fb_key --valuesize=fb_value --iadist=fb_ia --update=0.25 --depth=4 --measure_depth=1 --connections=16 --measure_connections=16 --measure_qps=2000 --qps=200000 --time=20")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil { panic(err) }
		time.Sleep(time.Second)

		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "stop"}
			}
		}
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "first"}
			}
		}
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "last"}
			}
		}
		for _, sheep := range(l_m.pasture) {
			for _, log := range(sheep.logs) {
				l_m.request_log_chan <- []string{sheep.id, log.id, "close"}
			}
		}

		var rep_sheep *sheep
		for _, sheep := range(l_m.pasture) {
			rep_sheep = sheep
			break
		}
		select {
		case bayopt_s.compute_ctrl_chan <- []string{l_m.id, rep_sheep.id}:
		default:
		}

		time.Sleep(time.Second)
	}

}

func bayopt_main(nodes []node) {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }

	// initialize generic shepherd
	s := shepherd{id: "sheperd-bayopt"}
	s.init(nodes)
	// initialize specialized energy-performance shepherd
	bayopt_s := bayopt_shepherd{shepherd:s,
				    logs_dir: home_dir + "/shepherd-bayopt-logs/"}
	bayopt_s.init()
	bayopt_s.init_local()
	
	// start all management and coordination threads
	bayopt_s.deploy_musters()
	go bayopt_s.listen_heartbeats()
	go bayopt_s.process_logs()
	go bayopt_s.compute_control()

	for _, l_m := range(bayopt_s.local_musters) {
		var ctrl_dvfs_id string
		var ctrl_itr_id string
		var ctrl_dvfs_val uint64
		var ctrl_itr_val uint64

		//starting optimization process..
		l_m.start_optimize_chan <- optimize_request{ntrials: 1}
		opt_rep := <- l_m.ready_optimize_chan
		ctrls := opt_rep.settings
		for _, ctrl := range(ctrls) {
			switch {
			case ctrl.knob == "dvfs":
				ctrl_dvfs_id = "dvfs-ctrl-"
				ctrl_dvfs_val = ctrl.val
			case ctrl.knob == "itr-delay":
				ctrl_itr_id = "itr-ctrl-" + l_m.ip
				ctrl_itr_val = ctrl.val
			default:
			}
		}
		//setting initial/start ctrl settings..
		for _, sheep := range(l_m.pasture) {
			c_str := strconv.Itoa(int(sheep.core))
			ctrl_dvfs_id = "dvfs-ctrl-" + c_str + "-" + l_m.ip
			start_ctrls := make(map[string]uint64)
			start_ctrls[ctrl_dvfs_id] = ctrl_dvfs_val
			start_ctrls[ctrl_itr_id] = ctrl_itr_val
			//applying ctrl..
			//TODO refactor
			l_m.new_ctrl_chan <- control_request{sheep_id: sheep.id, ctrls: start_ctrls}
			ctrl_reply := <- sheep.ready_ctrl_chan
			set_ctrls := ctrl_reply.ctrls
			done_ctrl := ctrl_reply.done
			if done_ctrl {
				for ctrl_id, ctrl_val := range(set_ctrls) {
				        sheep.controls[ctrl_id].value = ctrl_val
				}
			}
		}
		//running workload with optimization ready..
		go bayopt_s.run_workload(l_m.id)
	}


	for _, l_m := range(bayopt_s.local_musters) {
		for _, sheep := range(l_m.pasture) {
			for _, f := range(sheep.log_f_map) { defer f.Close() }
		}
	}
	time.Sleep(exp_timeout)
}





