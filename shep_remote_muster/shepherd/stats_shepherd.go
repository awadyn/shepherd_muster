package main

import (
	"os"
	"fmt"
	"slices"
	"sort"
)

/**************************************/
/****** SHEPHERD SPECIALIZATION  ******/
/**************************************/

// a stats_muster is an intlog_muster with additional data structures
// to store useful log statistics
type stats_muster struct {
	intlog_muster

	rx_bytes_all map[string][]uint64
	rx_bytes_median int
	processing_lock chan bool
}

type stats_shepherd struct {
	shepherd
	stats_musters map[string]*stats_muster
	rx_bytes_medians []int
}


func (stats_s *stats_shepherd) init() {
	home_dir, err := os.Getwd()
	if err != nil { panic(err) }
	logs_dir := home_dir + "/" + "stats-logs-"

	stats_s.stats_musters = make(map[string]*stats_muster)
	stats_s.rx_bytes_medians = make([]int, 0)

	for _, l_m := range(stats_s.local_musters) {
		l_m.logs_dir = logs_dir + l_m.id + "/"
		err := os.Mkdir(l_m.logs_dir, 0750)
		if err != nil && !os.IsExist(err) { panic(err) }

		intlog_m := intlog_muster{local_muster: *l_m}
		intlog_m.init()

		stats_m := stats_muster{intlog_muster: intlog_m}
		stats_m.init()

//		stats_m.logs_dir = logs_dir + stats_m.id + "/"
//		err := os.Mkdir(stats_m.logs_dir, 0750)
//		if err != nil && !os.IsExist(err) { panic(err) }

		stats_s.stats_musters[stats_m.id] = &stats_m
	}
}



/**************************/
/***** LOG PROCESSING *****/
/**************************/

/* 
   This function implements the log processing loop of a stats shepherd.
   Currently, this shepherd is specialized to 1) compute a metric's signal that
   concatenates this metric's signals from all request servicing cores (e.g. rx_bytes signal) 
   and 2) identify the signal by its median value.

   TODO ..
*/
func (stats_s stats_shepherd) process_logs(m_id string) {
	l_m := stats_s.stats_musters[m_id]
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
				if debug { fmt.Printf("\033[32m-------- SPECIALIZED PROCESS LOG SIGNAL :  %v - %v\n\033[0m", sheep_id, log_id) }

				mem_buff := *(log.mem_buff)
				// get per-sheep rx_bytes
				var rx_bytes_idx int = slices.Index(log.metrics, "rx_bytes")
				rx_bytes := make([]uint64, buff_max_size)
				j := 0
				for _, row := range(mem_buff) {
					rx_bytes[j] = row[rx_bytes_idx]
					j ++
				}

				<- l_m.processing_lock
				// append per-sheep rx_bytes to muster data 
				l_m.rx_bytes_all[sheep_id] = rx_bytes
//				fmt.Println(len(l_m.rx_bytes_all))
				// check if rx_bytes_all complete and compute rx_bytes_total
				if len(l_m.rx_bytes_all) == len(l_m.pasture) - 1 {
//					fmt.Println("READY TO GUESS QPS")
					rx_bytes_total := make([]int, 0, log.max_size)
					i := 0
					for {
						var total uint64 = 0
						for _, bytes := range(l_m.rx_bytes_all) {
							total += bytes[i]
						}
						// here if log entries ended
						if total == 0 { break }
						rx_bytes_total = append(rx_bytes_total, int(total))
						i += 1
					}
					sort.Ints(rx_bytes_total)
					l_m.rx_bytes_median = rx_bytes_total[len(rx_bytes_total)/2]
					fmt.Println("median: ", l_m.rx_bytes_median)
					l_m.rx_bytes_all = make(map[string][]uint64)
				}
				select {
				case l_m.processing_lock <- true:
				default:
				}
				// end new func

				if debug { fmt.Printf("\033[32m-------- COMPLETED SPECIALIZED PROCESS LOG :  %v - %v\n\033[0m", sheep.id, log.id) }

				select {
				case log.ready_process_chan <- true:
				default:
				}

			} ()
		}
	}
}







