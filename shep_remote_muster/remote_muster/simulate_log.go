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

func (r_m *remote_muster) simulate_remote_log(core uint8, log_id string) {
	fmt.Println(log_id,  " -- -- -- -- SIMULATING REMOTE LOG --")
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
				<- r_m.logs[log_id].ready_buff_chan
				return 
			}
			if err != nil { panic(err) }
			joules_val, _ := strconv.Atoi(row[0])
			timestamp_val, _ := strconv.Atoi(row[1])
			(*log.r_buff)[counter] = []uint64{uint64(joules_val), uint64(timestamp_val)}
			counter ++
		case counter == log.max_size:
			r_m.full_buff_chan <- log_id
			<- r_m.logs[log_id].ready_buff_chan
			*log.r_buff = make([][]uint64, log.max_size)
			counter = 0
		}
	}
}


