package main

import (
	"fmt"
	"os/exec"
	"strconv"
//	"time"
	"syscall"
)

/************************************/
/****** MUSTER SPECIALIZATION  ******/
/************************************/

type bayopt_muster struct {
	remote_muster
	logs_dir string
	ixgbe_metrics []string
	buff_max_size uint64
}

/* 
   read values of bayopt specialization control settings
   and initialize log files accordingly
*/
func (bayopt_m *bayopt_muster) init_remote() {
	iface := bayopt_m.get_internal_iface()
	if iface == "" {
		fmt.Printf("**** PROBLEM: %v cannot get internal network interface name.. aborting\n", bayopt_m.id)
		return
	}
	for _, sheep := range(bayopt_m.pasture) {
		for _, ctrl := range(sheep.controls) {
			switch {
			case ctrl.knob == "itr-delay":
				rx_usecs_reading := ctrl.getter(0, iface)	
				if rx_usecs_reading == 0 {
					fmt.Printf("**** PROBLEM: %v cannot read ethtool rx_usecs value.. assuming default value 1\n", bayopt_m.id)
					ctrl.value = 1
				} else {
					ctrl.value = rx_usecs_reading
				}
			case ctrl.knob == "dvfs":
				dvfs_reading := ctrl.getter(sheep.core)
				if dvfs_reading == 0 {
					fmt.Printf("**** PROBLEM: %v cannot read dvfs value.. assuming default value 0xffff\n", bayopt_m.id)
					ctrl.value = 0xffff
				} else {
					ctrl.value = dvfs_reading
				}
			default:
				fmt.Printf("**** PROBLEM: UNKOWN CTRL KNOB -- %v - %v - %v\n", bayopt_m.id, sheep.id, ctrl.id)
			}
		}
	}
	bayopt_m.init_log_files(bayopt_m.logs_dir)
}



/* The following functions associate each sheep with its
   per-interrupt log data in /proc/ixgbe_stats/core/sheep.core
*/
func ixgbe_native_log(sheep *sheep, log *log, logs_dir string) {
	c_str := strconv.Itoa(int(sheep.core))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := logs_dir + c_str

	cmd_flush := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd_flush.Run(); err != nil { 
		fmt.Printf("\033[31;1m****** PROBLEM: %v cannot attach to native logger.. aborting\n\033[0m", sheep.id)
		return
	}

	cmd := exec.Command("bash", "-c", "/users/awadyn/shepherd_muster/shep_remote_muster/read_ixgbe_stats.sh " + src_fname + " " + log_fname)
	cmd.SysProcAttr = &syscall.SysProcAttr{ Pdeathsig: syscall.SIGTERM }

	go func() {
		sheep := sheep
		cmd := cmd
		for {
			select {
			case <- sheep.detach_native_logger:
				err := cmd.Process.Kill()
				if err != nil { panic(err) }
//				fmt.Printf("\033[36;1m****** ALERT: killed native logger for %v\n\033[0m", sheep.id)
				return
			}
		}
	} ()

	if err := cmd.Run(); err != nil { 
//		fmt.Printf("\033[31;1m****** PROBLEM: %v cannot access native logger data.. aborting \n\033[0m", sheep.id)
		return
	}
}


func (bayopt_m *bayopt_muster) ctrl_manage(sheep_id string) {
	fmt.Printf("\033[36m-- MUSTER %v -- SHEEP %v - STARTING CONTROL MANAGER\n\033[0m", bayopt_m.id, sheep_id)
	sheep := bayopt_m.pasture[sheep_id]
	var err error
	for {
		select {
		case new_ctrls := <- sheep.new_ctrl_chan:
			for ctrl_id, ctrl_val := range(new_ctrls) {
				switch {
				case sheep.controls[ctrl_id].knob == "dvfs":
					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				case sheep.controls[ctrl_id].knob == "itr-delay":
					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				default:
				}
				if err != nil { panic(err) }
				sheep.controls[ctrl_id].value = ctrl_val
			}
			bayopt_m.pasture[sheep.id].ready_ctrl_chan <- control_reply{done: true, ctrls: new_ctrls}
		}
	}
}








