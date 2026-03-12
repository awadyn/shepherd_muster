package main

import (
	"strconv"
	"math"
	"slices"
	//"fmt"
)

/*********************************************/

type intlog_sheep struct {
	sheep

	rx_bytes []uint64
	timestamps []uint64

	processing_lock chan bool	
}

/* a load_pred_muster extends from an intlog_muster with additional data structures 
   for storing log statistics to use in load prediction and control optimization
*/
type load_pred_muster struct {
	intlog_muster
	intlog_pasture map[string]*intlog_sheep

	processing_lock chan bool	// write protection if multiple sheep access muster structs
	
	rx_bytes_concat []uint64
	rx_bytes_median uint64

	ctrl_break uint16
	cur_load uint32
	cur_load_guess uint32
}

/* a load prediction muster manages one Memcached server
   it holds data to maintain statistics about the load imposed on the server
   it issues control state change requests when that load changes
   the control state managed in this case is the interrupt-delay and dvfs
   setting of the full node 
*/
func (load_pred_m *load_pred_muster) init() {
	load_pred_m.processing_lock = make(chan bool, 1)
	load_pred_m.processing_lock <- true
	
	load_pred_m.rx_bytes_concat = make([]uint64, 0)
	load_pred_m.ctrl_break = 1
	load_pred_m.cur_load = 0
	load_pred_m.cur_load_guess = 0

	for sheep_id, sheep := range(load_pred_m.pasture) {
		if (sheep.label != "node") { continue }
		index := strconv.Itoa(int(sheep.index))
		label := sheep.label
		ctrl_itr := control{id: "itr-ctrl-" + label + "-" + index + "-" + load_pred_m.ip, n_ip: load_pred_m.ip}
		ctrl_itr.init("itr-delay", read_rx_usecs, write_rx_usecs)
		ctrl_dvfs := control{id: "dvfs-ctrl-" + label + "-" + index + "-" + load_pred_m.ip, n_ip: load_pred_m.ip}
		ctrl_dvfs.init("dvfs", read_dvfs_all, write_dvfs_all)
		load_pred_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
		load_pred_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
	}

	// specialize sheep
	intlog_pasture := make(map[string]*intlog_sheep)
	for sheep_id, sheep := range(load_pred_m.pasture) {
		rx_bytes := make([]uint64, 0)
		timestamps := make([]uint64, 0)
		intlog_sh := intlog_sheep{sheep: *sheep, rx_bytes: rx_bytes, timestamps: timestamps}
		intlog_sh.processing_lock = make(chan bool, 1)
		intlog_sh.processing_lock <- true
		intlog_pasture[sheep_id] = &intlog_sh
	}
	load_pred_m.intlog_pasture = intlog_pasture

	// TODO delete muster pasture if possible
}


func (load_pred_m *load_pred_muster) rx_to_load_guess() uint32 {
	var diffs map[uint32]float64 = make(map[uint32]float64)
	var cur_itrd uint16 = 100 // TODO fix
	qps_medians := itrd_qps_med_map[cur_itrd]
	for qps, med := range(qps_medians) {
		diffs[qps] = math.Abs(float64(load_pred_m.rx_bytes_median) - float64(med))
	}
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


func (load_pred_m *load_pred_muster) get_rx_signal(intlog_sh *intlog_sheep, l log) {
	mem_buff := l.mem_buff
	var rx_bytes_idx int = slices.Index(l.metrics, "rx_bytes")
	var timestamp_idx int = slices.Index(l.metrics, "timestamp")
	timestamps := make([]uint64, len(*mem_buff))
	rx_bytes := make([]uint64, len(*mem_buff))
	for i := 0; i < len(*mem_buff); i ++ {
		rx_bytes[i] = (*mem_buff)[i][rx_bytes_idx]
		timestamps[i] = (*mem_buff)[i][timestamp_idx]
	}

	<- intlog_sh.processing_lock
	intlog_sh.rx_bytes = append(intlog_sh.rx_bytes, rx_bytes...)
	intlog_sh.timestamps = append(intlog_sh.timestamps, timestamps...)
	intlog_sh.processing_lock <- true
}

func (load_pred_m *load_pred_muster) check_ready_pred() bool {
	ready := true
	for _, sheep := range(load_pred_m.intlog_pasture) {
		if sheep.label == "node" { continue }
		<- sheep.processing_lock
		if len(sheep.rx_bytes) < 1024 { ready = false }
		sheep.processing_lock <- true
	}
	return ready 
}


func (load_pred_m *load_pred_muster) concat_rx_bytes() []int {
	iterators := make(map[string]int)
	total_length := 0
	for sheep_id, sheep := range(load_pred_m.intlog_pasture) { 
		if sheep.label == "node" { continue }
		iterators[sheep_id] = 0 
		total_length += len(sheep.rx_bytes)
	}
	load_pred_m.rx_bytes_concat = make([]uint64, total_length)
	ints_concat := make([]int, total_length)
	min_timestamp := uint64(math.Pow(2, 64))
	ref_sheep := ""
	var ref_rx_bytes uint64 = 0
	for i := 0; i < total_length; i ++ {
		for sheep_id, sheep := range(load_pred_m.intlog_pasture) {
			if sheep.label == "node" { continue }
			sheep_itr := iterators[sheep_id]
			if sheep_itr == len(sheep.timestamps) { continue }
			if sheep.timestamps[sheep_itr] <= min_timestamp {
				min_timestamp = sheep.timestamps[sheep_itr]
				ref_rx_bytes = sheep.rx_bytes[sheep_itr]
				ref_sheep = sheep_id
			}
		}
		load_pred_m.rx_bytes_concat[i] = ref_rx_bytes
		ints_concat[i] = int(ref_rx_bytes)
		iterators[ref_sheep] += 1
		min_timestamp = uint64(math.Pow(2, 64))
	}
	return ints_concat
}




