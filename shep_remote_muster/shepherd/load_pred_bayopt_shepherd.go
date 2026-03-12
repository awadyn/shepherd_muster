package main

import ( 
	"fmt"
	"strconv"
//	"time"
//	"math"
//	"slices"
	"sort"
//	"google.golang.org/protobuf/types/known/anypb"
//	"google.golang.org/protobuf/types/known/wrapperspb"
)



/* a load_pred_bayopt_shepherd extends from an intlog_shepherd and has additional functionality
   for load prediction and control configuration via Bayesian optimization
*/
type load_pred_bayopt_shepherd struct {
	intlog_shepherd
	load_pred_musters map[string]*load_pred_muster
}



func (load_pred_bayopt_s *load_pred_bayopt_shepherd) init() {
	load_pred_bayopt_s.load_pred_musters = make(map[string]*load_pred_muster)

	for _, intlog_m := range(load_pred_bayopt_s.intlog_musters) {
		load_pred_m := load_pred_muster{intlog_muster: *intlog_m}
		load_pred_m.init()
		load_pred_bayopt_s.load_pred_musters[intlog_m.id] = &load_pred_m
	}
}



func (load_pred_bayopt_s *load_pred_bayopt_shepherd) start_exp() {
	// start log coordination with intloggers on all muster nodes
	load_pred_bayopt_s.intlog_shepherd.start_exp()
	
	// start optimization coordination with log processing and/or bayesian optimizer servers
	// TODO
}

func (load_pred_bayopt_s *load_pred_bayopt_shepherd) check_ready_ctrl(m_id string, guess uint32) {
	l_m := load_pred_bayopt_s.load_pred_musters[m_id]
	if (guess != l_m.cur_load_guess) {	// new load predicted
		l_m.cur_load_guess = guess	// reset load prediction
		l_m.ctrl_break = 1		// reset count N
	} else {					// same load predicted
		if (l_m.ctrl_break == 0) {	// consistent load prediction N times

			ctrl_req := make([]any, 0)
			ctrl_req = append(ctrl_req, guess)
			l_m.process_ctrl_chan <- ctrl_req
			<- l_m.done_ctrl_chan
			l_m.cur_load = guess	// update load to predicted load value
			l_m.ctrl_break = 1	// reset count N
		} else {				// not ready yet to apply control
			l_m.ctrl_break = (l_m.ctrl_break + 1) % 3
		}
	}	
}


func (load_pred_bayopt_s *load_pred_bayopt_shepherd) pred_load(m_id string) {
	l_m := load_pred_bayopt_s.load_pred_musters[m_id]
	for _, sheep := range(l_m.intlog_pasture) {
		if sheep.label == "node" { continue }
		<- sheep.processing_lock
	}
	rx_bytes_concat := l_m.concat_rx_bytes()
	sort.Ints(rx_bytes_concat)
	l_m.rx_bytes_median = uint64(rx_bytes_concat[len(rx_bytes_concat)/2])
	guess := l_m.rx_to_load_guess()
	fmt.Println("****** QPS GUESS -- ", guess, " -- MEDIAN -- ", l_m.rx_bytes_median)

	load_pred_bayopt_s.check_ready_ctrl(m_id, guess)

	for _, sheep := range(l_m.intlog_pasture) {	// reset data structs
		if sheep.label == "node" { continue }
		sheep.timestamps = make([]uint64, 0)
		sheep.rx_bytes = make([]uint64, 0)
	}
	l_m.rx_bytes_concat = make([]uint64, 0)
	for _, sheep := range(l_m.intlog_pasture) {
		if sheep.label == "node" { continue }
		sheep.processing_lock <- true
	}
}



func (load_pred_bayopt_s load_pred_bayopt_shepherd) process_logs(m_id string) {
	l_m := load_pred_bayopt_s.load_pred_musters[m_id]
	for {
		select {
		case ids := <- l_m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			sheep := l_m.intlog_pasture[sheep_id]
			log := *(sheep.logs[log_id])

			if debug { fmt.Printf("\033[32m------------ SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep.id, log.id) }
			l_m.get_rx_signal(sheep, log)
			log.ready_process_chan <- true /* can overwrite mem_buff now */
			ready := l_m.check_ready_pred()
			if (ready) { go load_pred_bayopt_s.pred_load(m_id) }
			if debug { fmt.Printf("\033[32m------------ COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }
			
		}
	}
}

func (load_pred_bayopt_s load_pred_bayopt_shepherd) process_control(m_id string) {
	l_m := load_pred_bayopt_s.load_pred_musters[m_id]
	for {
		select {
		case req := <- l_m.process_ctrl_chan:
			fmt.Println(req)
			qps_guess := req[0].(uint32)
			itrd_val, _ := strconv.ParseUint(opt_itrd[qps_guess], 10, 64)
			dvfs_val, _ := strconv.ParseUint(opt_dvfs[qps_guess], 16, 64)
			new_ctrls := make(map[string]uint64)
			for _, sheep := range(l_m.intlog_pasture) {
				if sheep.label != "node" { continue }
				fmt.Println("*********** APPLYING CTRLS ***************", sheep.id, new_ctrls)
				index := strconv.Itoa(int(sheep.index))
				label := sheep.label
				new_ctrls["itr-ctrl-" + label + "-" + index + "-" + l_m.ip] = itrd_val
				new_ctrls["dvfs-ctrl-" + label + "-" + index + "-" + l_m.ip] = dvfs_val
				l_m.new_ctrl_chan <- control_request{sheep_id: sheep.id, ctrls: new_ctrls}
				ctrl_reply := <- sheep.done_ctrl_chan
				fmt.Println("*********** DONE CTRLS ***************", sheep.id, ctrl_reply)
			}
			l_m.done_ctrl_chan <- true
		}
	}

}


