package main

import (
	"os"
	"os/exec"
	"strconv"
	"fmt"
	"time"
	"strings"
)


func parseProcEntry(data []byte, log *log) {
	for _, line := range strings.Split(string(data), "\n") {
		if len(line) == 0 {continue}
		parts := strings.Fields(line)
		int_parts := make([]uint64, len(parts))
		for i, val := range(parts) {
			int_val, err := strconv.ParseUint(val, 10, 64)
			if err != nil { 
				fmt.Printf("ERROR PARSING PROC ENTRY %v %v %v %v\n", i, val, log.id, err)
				continue 
			} else {
				int_parts[i] = int_val
			}
		}
		*log.log_buff = append(*log.log_buff, int_parts)
	}
}

func readProcEntry(sheep *sheep, log *log) {
	fmt.Println("HERE " + log.id)
	coreID := sheep.index
	path := fmt.Sprintf("/proc/ixgbe_stats/core/%d", coreID)
	for {
		select {
		case <- sheep.detach_native_logger:
			return
		default:
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("reading core %d stats: %w\n", err)
				return
			}
			parseProcEntry(data, log)

			fmt.Printf("core %d - len %d - len(log_buff) %d\n", coreID, len(data), len(*log.log_buff))
			time.Sleep(time.Second * 6)
		}
	}
}


/* The following functions associate each sheep with its
   per-interrupt log data in /proc/ixgbe_stats/core/sheep.core
*/
func ixgbe_native_log(sheep *sheep, log *log, logs_dir string) {
	if sheep.label != "core" { return }
	c_str := strconv.Itoa(int(sheep.index))
	src_fname := "/proc/ixgbe_stats/core/" + c_str

	cmd_flush := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd_flush.Run(); err != nil { 
		fmt.Printf("\033[31;1m****** PROBLEM: %v cannot attach to native logger.. aborting\n\033[0m", sheep.id)
		return
	}

	go readProcEntry(sheep, log)
}



