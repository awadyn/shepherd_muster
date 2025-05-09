package main

import (
	"os/exec"
	"strconv"
	"fmt"
	"syscall"
)

/* The following functions associate each sheep with its
   per-interrupt log data in /proc/ixgbe_stats/core/sheep.core
*/
func ixgbe_native_log(sheep *sheep, log *log, logs_dir string) {
	if sheep.label != "core" { return }
	c_str := strconv.Itoa(int(sheep.index))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := logs_dir + sheep.label + "-" + strconv.Itoa(int(sheep.index))

	cmd_flush := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd_flush.Run(); err != nil { 
		fmt.Printf("\033[31;1m****** PROBLEM: %v cannot attach to native logger.. aborting\n\033[0m", sheep.id)
		return
	}

	cmd := exec.Command("bash", "-c", "/users/awadyn/shepherd_muster/shep_remote_muster/read_ixgbe_stats.sh " + src_fname + " " + log_fname)
	cmd.SysProcAttr = &syscall.SysProcAttr{ Pdeathsig: syscall.SIGTERM }

	go func() {
		cmd := cmd
		if err := cmd.Run(); err != nil { 
			return
		}
	} ()

	go func() {
		sheep := sheep
		cmd := cmd
		for {
			select {
			case <- sheep.detach_native_logger:
				err := cmd.Process.Kill()
				if err != nil { panic(err) }
				return
			}
		}
	} ()
}



