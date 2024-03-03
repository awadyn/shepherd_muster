package main

import (
	"fmt"
//	"time"
	"strconv"
	"os"
	"encoding/csv"
	"io"
)

//var flink_cols = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
//		"instructions", "cycles", "ref_cycles", "llc_miss", 
//		"c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}

var mcd_cols = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
		"instructions", "cycles", "ref_cycles", "llc_miss", 
		"c1", "c1e", "c3", "c6", "c7", "joules","timestamp"}

func (r_m *remote_muster) simulate_remote_log(core uint8, log_id string) {
	fmt.Println(log_id,  " -- -- -- -- SIMULATING REMOTE LOG --")

//	log_fname := "/home/tanneen/shepherd_muster/flink_mapper_itr_logs/linux.flink.dmesg._" + strconv.Itoa(int(core)) + "_0"
	log_fname := "/home/tanneen/shepherd_muster/mcd_logs/linux.mcd.dmesg.0_" + strconv.Itoa(int(core)) + "_100_0x1100_135_200000.ep.csv"
	fmt.Println(log_id, " -- -- -- -- LOG FILE -- ", log_fname)
	f, err := os.Open(log_fname)
	if err != nil { panic(err) }
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ' '

	var counter uint64 
	counter = 0
	log := r_m.logs[log_id]
	for {
		switch {
		case counter < log.max_size:
			row, err := reader.Read()
			if err == io.EOF { 
				fmt.Println(log.r_buff)
				r_m.full_buff_chan <- log_id
				return 
			}
			if err != nil { panic(err) }
			joules_val, _ := strconv.Atoi(row[0])
			timestamp_val, _ := strconv.Atoi(row[1])
			(*log.r_buff)[counter] = []uint64{uint64(joules_val), uint64(timestamp_val)}
			counter ++
		case counter == log.max_size:
			r_m.full_buff_chan <- log_id
			<- r_m.logs[log_id].ready_chan
			*log.r_buff = make([][]uint64, log.max_size)
			counter = 0
		}
	}

//	for {
//		select {
//		case log_id := <- m.kill_log:
//			fmt.Println(log_id, " -- -- -- -- KILLING REMOTE LOG  --")
//			log.mem = make([][]uint64, 0)
//			go m.simulate_remote_log(log_id, c, v)
//			return
//		default:
//			if counter > log.max_size { 
//			}
//		}
//	}
}


