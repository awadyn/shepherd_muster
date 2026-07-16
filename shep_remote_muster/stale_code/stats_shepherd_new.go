package main

import (
	"fmt"
	"slices"
	"sort"
	"math"
	"strconv"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

// a stats_muster is an intlog_muster with additional data structures
// to store useful log statistics
type stats_muster struct {
	intlog_muster

	rx_bytes_all map[string][]uint64
	timestamps_all map[string][]uint64
	rx_bytes_concat []int
	rx_bytes_medians []int
	rx_bytes_median int
	processing_lock chan bool
}

type stats_shepherd struct {
	intlog_shepherd
	stats_musters map[string]*stats_muster
	rx_bytes_medians map[int][]int
}


func (stats_s *stats_shepherd) init() {
	stats_s.stats_musters = make(map[string]*stats_muster)
	stats_s.rx_bytes_medians = make(map[int][]int)

	for _, intlog_m := range(stats_s.intlog_musters) {
		stats_m := stats_muster{intlog_muster: *intlog_m}
		stats_m.init()
		stats_s.stats_musters[stats_m.id] = &stats_m
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

func (stats_s stats_shepherd) concat_rx_bytes(m_id string) {
	l_m := stats_s.stats_musters[m_id]

	iterators := make(map[string]int)
	total_length := 0
	for sheep_id, sheep := range(l_m.pasture) { 
		if sheep.label == "node" { continue }
		iterators[sheep_id] = 0 
		total_length += len(l_m.rx_bytes_all[sheep_id])
	}

	l_m.rx_bytes_concat = make([]int, total_length)
	min_timestamp := uint64(math.Pow(2, 64))
	ref_iterator := 0
	ref_sheep := ""
	for j := 0; j < total_length; j ++ {
		for sheep_id, sheep := range(l_m.pasture) {
			if sheep.label == "node" { continue }
			sheep_itr := iterators[sheep_id]
			sheep_timestamps := l_m.timestamps_all[sheep_id]
			if sheep_itr == len(sheep_timestamps) { continue }
			if sheep_timestamps[sheep_itr] <= min_timestamp {
				min_timestamp = sheep_timestamps[sheep_itr]
				ref_iterator = sheep_itr
				ref_sheep = sheep_id
			}
		}
		l_m.rx_bytes_concat[j] = int(l_m.rx_bytes_all[ref_sheep][ref_iterator])
		iterators[ref_sheep] += 1
		min_timestamp = uint64(math.Pow(2, 64))
	}


}

/* 
   This function implements the log processing loop of a stats shepherd.
   Currently, this shepherd is specialized to 1) compute a metric's signal that
   concatenates this metric's signals from all request servicing cores (e.g. rx_bytes signal) 
   and 2) identify the signal by its median value.

   TODO ..
*/
func (stats_s stats_shepherd) process_logs(m_id string) {
	l_m := stats_s.stats_musters[m_id]
	ctrl_break := 1
	cur_qps_guess := -1
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
				if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep.id, log.id) }

				rx_bytes, timestamps := stats_s.get_rx_signal(log)

				<- l_m.processing_lock
				// append per-sheep rx_bytes signal to stats_muster data map
				l_m.timestamps_all[sheep_id] = append(l_m.timestamps_all[sheep_id], timestamps...)
				l_m.rx_bytes_all[sheep_id] = append(l_m.rx_bytes_all[sheep_id], rx_bytes...)

				// check if rx_bytes_all complete and concat
				ready := true
				if len(l_m.rx_bytes_all) == len(l_m.pasture) - 1 {
					for sheep_id, sheep := range(l_m.pasture) {
						if sheep.label == "node" { continue }
						if len(l_m.rx_bytes_all[sheep_id]) < 2048 { ready = false }
					}
					if ready {
						stats_s.concat_rx_bytes(l_m.id)
						
						// get current median of the concatenated rx-bytes frame
						sort.Ints(l_m.rx_bytes_concat)
						l_m.rx_bytes_median = l_m.rx_bytes_concat[len(l_m.rx_bytes_concat)/2]

						// OFFLINE PHASE: append to medians list of per-frame medians across run
						//l_m.rx_bytes_medians = append(l_m.rx_bytes_medians, l_m.rx_bytes_median)

						/* ONLINE PHASE: testing with manually collected medians */
						var cur_itrd uint64
						for _, sheep := range(l_m.pasture) {
							if sheep.label == "node" {
								cur_ctrls := sheep.controls
								for _, ctrl := range(cur_ctrls) {
									if ctrl.knob == "itr-delay" {
										cur_itrd = ctrl.value
										break
									}
								}
								break
							}
						}

						// big hack
						if cur_itrd == 0 { cur_itrd = 1 }
						fmt.Println(cur_itrd)

						var diffs map[uint32]uint64 = make(map[uint32]uint64)
						for itrd, qps_medians := range(itrd_qps_med_map) {
							if uint64(itrd) == cur_itrd {
								fmt.Println(qps_medians)
								for qps, med := range(qps_medians) {
									diffs[qps] = uint64(math.Abs(float64(l_m.rx_bytes_median - int(med))))
								}
							}
						}

//						for q := 0; q < len(qpses); q ++ {
//							diffs[q] = math.Abs(float64(l_m.rx_bytes_median - medians[q]))
//						}

						var guess uint32
						var min_diff uint64 = uint64(math.Pow(2, 64))
						for qps, diff := range(diffs) {
							if diff < min_diff { 
								min_diff = diff
								guess = qps
							}

						}

//						for q := 0; q < len(qpses); q ++ {
//							if diffs[q] < min_diff { 
//								min_diff = diffs[q] 
//								guess = q
//							}
//						}

						fmt.Println("************ QPS GUESS -- ", guess, " -- MEDIAN -- ", l_m.rx_bytes_median)
						if cur_qps_guess == int(guess) { 
							ctrl_break = 1 
						} else {
							if ctrl_break == 0 {
								fmt.Println("************ APPLYING CTRLS **********************"/*, opt_dvfs[guess], opt_itrd[guess]*/)
								new_ctrls := make(map[string]uint64)
								for _, sheep := range(stats_s.musters[m_id].pasture) {
									if sheep.label == "node" {
										index := strconv.Itoa(int(sheep.index))
										label := sheep.label
										itrd_val, _ := strconv.ParseUint(opt_itrd[guess], 10, 64)
										dvfs_val, _ := strconv.ParseUint(opt_dvfs[guess], 16, 64)
										new_ctrls["itr-ctrl-" + label + "-" + index + "-" + l_m.ip] = itrd_val
										new_ctrls["dvfs-ctrl-" + label + "-" + index + "-" + l_m.ip] = dvfs_val
										stats_s.control(m_id, sheep.id, new_ctrls)
										break
									}
								}
								cur_qps_guess = int(guess)
							}
							ctrl_break = (ctrl_break + 1) % 3
						}

//						fmt.Println("ref_sheep: ", ref_sheep, len(l_m.rx_bytes_concat),  "max size:", max_size, "min size:", min_size, "rx_bytes_median: ", l_m.rx_bytes_median)

						// reset/cleanup
						l_m.timestamps_all = make(map[string][]uint64)
						l_m.rx_bytes_all = make(map[string][]uint64)
						for _, sheep := range(l_m.pasture) {
							if sheep.label == "node" { continue }
							l_m.timestamps_all[sheep.id] = make([]uint64, 0)
							l_m.rx_bytes_all[sheep.id] = make([]uint64, 0)
						}
						l_m.rx_bytes_concat = make([]int, 0)
					}
				}
				select {
				case l_m.processing_lock <- true:
				default:
				}

				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }

				select {
				case log.ready_process_chan <- true:
				default:
				}
			} ()
		}
	}
}

func (stats_s stats_shepherd) get_rx_signal(l log) ([]uint64, []uint64) {
	// get per-sheep rx_bytes
	mem_buff := l.mem_buff
	var rx_bytes_idx int = slices.Index(l.metrics, "rx_bytes")
	var timestamp_idx int = slices.Index(l.metrics, "timestamp")
	timestamps := make([]uint64, len(*mem_buff))
	rx_bytes := make([]uint64, len(*mem_buff))
	for j := 0; j < len(*mem_buff); j ++ {
		rx_bytes[j] = (*mem_buff)[j][rx_bytes_idx]
		timestamps[j] = (*mem_buff)[j][timestamp_idx]
	}
	return rx_bytes, timestamps
}







