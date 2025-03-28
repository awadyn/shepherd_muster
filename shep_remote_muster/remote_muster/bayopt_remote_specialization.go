package main

import (
	"fmt"
)

/************************************/
/****** MUSTER SPECIALIZATION  ******/
/************************************/

type bayopt_muster struct {
	remote_muster
	ixgbe_metrics []string
	buff_max_size uint64
}

/* 
   read values of bayopt specialization control settings
   and initialize log files accordingly
*/
func (bayopt_m *bayopt_muster) init_remote() {
	iface := get_internal_iface()
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












