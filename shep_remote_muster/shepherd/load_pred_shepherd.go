package main

import (
	"fmt"
	"strconv"
	"flag"
	"math"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

/* a load_pred_muster extends from an intlog_muster with additional data structures used
   during the process of log processing and load prediction
   to store useful log statistics
*/
type load_pred_muster struct {
	intlog_muster
	
	// write protection in case of multiple per-sheep threads accessing muster datatypes
	processing_lock chan bool		
	
	rx_bytes_all map[string][]uint64
	timestamps_all map[string][]uint64
	rx_bytes_concat []uint64
	rx_bytes_median uint64

	ctrl_break uint16
	cur_load_pred uint32
	cur_load_guess uint32
}

/* a load_pred_shepherd extends from an intlog_shepherd with additional log processing
   functionality that enables load prediction
*/
type load_pred_shepherd struct {
	intlog_shepherd
	load_pred_musters map[string]*load_pred_muster
}


func (load_pred_s *load_pred_shepherd) init() {
	load_pred_s.load_pred_musters = make(map[string]*load_pred_muster)

	for _, intlog_m := range(load_pred_s.intlog_musters) {
		load_pred_m := load_pred_muster{intlog_muster: *intlog_m}
		load_pred_m.init()
		load_pred_s.load_pred_musters[intlog_m.id] = &load_pred_m
	}
}




func (load_pred_s *load_pred_shepherd) init_optimizers(server_ports []uint64, client_ports []uint64) {
	load_pred_s.optimizers = make(map[string]*optimizer)
	log_proc_optimizer := optimizer{id: "log_proc_opt"}
	for i, p := range(server_ports) {
		if (i == 0) {
			i_str := strconv.Itoa(int(i))
			port := flag.Int("optimize_server_port_" + i_str + "_" + load_pred_s.id, int(p), 
					 "local muster optimization server port")
			log_proc_optimizer.port = port
		}
	}
	for i, p := range(client_ports) {
		if (i == 0) {
			i_str := strconv.Itoa(int(i))
			addr := flag.String("optimize_server_addr_" + i_str + "_" + load_pred_s.id,  
							"localhost:" + strconv.Itoa(int(p)),
							"address of optimization client")
			log_proc_optimizer.addr = addr
		}
	}
	load_pred_s.optimizers["log_proc_opt"] = &log_proc_optimizer
	load_pred_s.optimizers["log_proc_opt"].start_optimize_chan = make(chan start_optimize_request, 1)
	load_pred_s.optimizers["log_proc_opt"].ready_optimize_chan = make(chan bool, 1)
	load_pred_s.optimizers["log_proc_opt"].request_optimize_chan = make(chan optimize_request, 1)
	load_pred_s.optimizers["log_proc_opt"].done_optimize_chan = make(chan bool, 1)
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

func (load_pred_s *load_pred_shepherd) load_pred() uint32 {
	var cur_itrd uint64 = 100 

	var diffs map[uint32]float64 = make(map[uint32]float64)
	for itrd, qps_medians := range(itrd_qps_med_map) {
		if uint64(itrd) == cur_itrd {
			for qps, med := range(qps_medians) {
				diffs[qps] = math.Abs(float64(rx_bytes_median) - float64(med))
			}
		}
	}
//	fmt.Println("QPS median diffs: ", diffs)

	var guess uint32
	var min_diff float64 = math.Pow(2, 64)
	for qps, diff := range(diffs) {
		if diff < min_diff { 
			min_diff = diff
			guess = qps
		}
	}

	return guess
}



func (load_pred_s load_pred_shepherd) process_logs(m_id string) {
	l_m := load_pred_s.load_pred_musters[m_id]
	var cur_guess uint32 = 0
	for {
		select {
		case ids := <- l_m.process_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			sheep := l_m.pasture[sheep_id]
			log := *(sheep.logs[log_id])
			go func() {
				sheep := sheep
				log := log
				sheep_id := sheep_id
				log_id := log_id
				if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep_id, log_id) }
		
				opt_args := make([]*anypb.Any, 0)
				data := wrapperspb.String(log.id)
				arg, _ := anypb.New(data)
				opt_args = append(opt_args, arg)
				optimizer := load_pred_s.optimizers["log_proc_opt"]
				optimizer.request_optimize_chan <- optimize_request{args: opt_args}
				done := <- optimizer.done_optimize_chan
				if !done {
					fmt.Printf("** ** ** ERROR ** ** ** OPTIMIZER %v COULD NOT PROCESS LOG\n", optimizer.id)
				}
				
				guess := load_pred_s.load_pred()
				if guess != cur_guess {
					fmt.Println("************ QPS GUESS -- ", guess, " -- MEDIAN -- ", rx_bytes_median)
					cur_guess = guess
				}

				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }
				log.ready_process_chan <- true
			} ()
		}
	}
}


//func (load_pred_s load_pred_shepherd) process_logs(m_id string) {
//	l_m := load_pred_s.load_pred_musters[m_id]
//	for {
//		select {
//		case ids := <- l_m.process_buff_chan:
//			sheep_id := ids[0]
//			log_id := ids[1]
//			sheep := l_m.pasture[sheep_id]
//			log := *(sheep.logs[log_id])
//			go func() {
//				sheep := sheep
//				log := log
//				l_m := l_m
//				log.log_file_lock <- true
//
//				// rpc to python log processing server with sheep_id, log_id, logs_dir
//				// signal done immediately 
//				// log processing server may return after some condition passes
//				// in this case, it will rpc to the shepherd who will signal process_control with the returned stats
//
//
//				rx_bytes, timestamps := get_rx_signal(log)
//
//				<- l_m.processing_lock
//
//				// append per-sheep rx_bytes signal to map of per-sheep rx_bytes signals
//				l_m.timestamps_all[sheep.id] = append(l_m.timestamps_all[sheep.id], timestamps...)
//				l_m.rx_bytes_all[sheep.id] = append(l_m.rx_bytes_all[sheep.id], rx_bytes...)
//
//				ready := l_m.is_ready_load_pred()
//
//				if ready {
//					l_m.compute_rx_median()
//					guess := l_m.load_pred()
//					if l_m.cur_load_guess == 0 { l_m.cur_load_guess = guess }
//					fmt.Println("************ QPS GUESS -- ", guess, " -- CURRENT QPS GUESS -- ", l_m.cur_load_pred, " -- MEDIAN -- ", l_m.rx_bytes_median)
//
//					load_pred_s.check_ready_control(l_m.id, guess)
//
//					// reset/cleanup
//					l_m.timestamps_all = make(map[string][]uint64)
//					l_m.rx_bytes_all = make(map[string][]uint64)
//					l_m.rx_bytes_concat = make([]uint64, 0)
//				}
//
//				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }
//				select {
//				case l_m.processing_lock <- true:
//				default:
//				}
//				select {
//				case log.ready_process_chan <- true:
//				default:
//				}
//			} ()
//		}
//	}
//}


func (load_pred_s load_pred_shepherd) process_control() {
//	for {
//		select {
//		case new_muster_ctrls := <- load_pred_s.new_ctrl_chan:
//			for m_id, ctrl_req := range(new_muster_ctrls) {
//				fmt.Println(m_id, ctrl_req)
//				for sheep_id, ctrls := range(ctrl_req) {
//					fmt.Println(sheep_id, ctrls)
//				}
//			}
//		}
//	}
}




