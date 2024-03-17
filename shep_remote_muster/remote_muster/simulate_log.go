package main

import (
	"fmt"
//	"time"
	"strconv"
	"os"
	"encoding/csv"
	"io"
)

var mcd_cols = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
		"instructions", "cycles", "ref_cycles", "llc_miss", 
		"c1", "c1e", "c3", "c6", "c7", "joules","timestamp"}

func (r_m *remote_muster) simulate_remote_log(sheep_id string, log_id string, core uint8) {
	c_str := strconv.Itoa(int(core))

	log := r_m.pasture[sheep_id].logs[log_id]
	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + r_m.ip
	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + r_m.ip

	dvfs_val := r_m.pasture[sheep_id].controls[ctrl_dvfs_id].value
	itr_val := r_m.pasture[sheep_id].controls[ctrl_itr_id].value
	dvfs_str := fmt.Sprintf("0x%x", dvfs_val)
	itr_str := strconv.Itoa(int(itr_val))
	log_fname := "/home/tanneen/shepherd_muster/mcd_logs/linux.mcd.dmesg.0_" + c_str + "_" + itr_str + "_" + dvfs_str + "_135_200000.ep.csv.ep"

	fmt.Printf("-- -- CORE %v -- -- SIMULATING REMOTE LOG :  %v \n", c_str, log_fname)
	f, err := os.Open(log_fname)
	if err != nil { panic(err) }
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = ' '

	var counter uint64 
	counter = 0
	done := false
	for {
		if done { break }
		switch {
		case counter < log.max_size:
			row, err := reader.Read()
			if err == io.EOF { 
				r_m.full_buff_chan <- []string{sheep_id, log_id}
				<- log.ready_buff_chan
				done = true 
				break
			}
			if err != nil { panic(err) }
			joules_val, _ := strconv.Atoi(row[0])
			timestamp_val, _ := strconv.Atoi(row[1])
			(*log.r_buff)[counter] = []uint64{uint64(joules_val), uint64(timestamp_val)}
			counter ++
		case counter == log.max_size:
			r_m.full_buff_chan <- []string{sheep_id, log_id}
			<- log.ready_buff_chan
			*log.r_buff = make([][]uint64, log.max_size)
			counter = 0
		}
	}
	r_m.pasture[sheep_id].logs[log_id].done_chan <- true
}


