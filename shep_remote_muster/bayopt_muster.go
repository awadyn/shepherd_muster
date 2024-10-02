package main

import (
	"fmt"
//	"os"
	"os/exec"
	"strconv"
//	"encoding/csv"
//	"time"
//	"io"
	"strings"
	"bytes"
//	"slices"
//	"bufio"
)

/*********************************************/


func (bayopt_m *bayopt_muster) init() {
	bayopt_m.ixgbe_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	bayopt_m.buff_max_size = 1

	ctrl_itr_shared := control{id: "itr-ctrl-" + bayopt_m.ip, n_ip: bayopt_m.ip}
	ctrl_itr_shared.init("itr-delay", read_rx_usecs, write_rx_usecs)
	for sheep_id, sheep := range(bayopt_m.pasture) {
		c_str := strconv.Itoa(int(sheep.core))
		log_id := "ixgbe-log-" + c_str + "-" + bayopt_m.ip
		log_c := log{id: log_id}
		log_c.init(bayopt_m.buff_max_size, bayopt_m.ixgbe_metrics, 3)
		ctrl_dvfs := control{id: "dvfs-ctrl-" + c_str + "-" + bayopt_m.ip, n_ip: bayopt_m.ip}
		ctrl_dvfs.init("dvfs", read_dvfs, write_dvfs)
		bayopt_m.pasture[sheep_id].logs[log_id] = &log_c	
		bayopt_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		bayopt_m.pasture[sheep_id].controls[ctrl_itr_shared.id] = &ctrl_itr_shared
	}
}

func (bayopt_m *bayopt_muster) get_internal_iface() string {
	var out strings.Builder
	var stderr strings.Builder
	cmd:= exec.Command("bash", "-c", "ls /sys/class/net | grep enp | grep f0")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	iface := out.String()
	stderr_str := stderr.String()
	if len(iface) == 0 {
		if len(stderr_str) > 0 {
			fmt.Printf("**** PROBLEM: %v cannot read ethernet interface name.. aborting..\n", bayopt_m.id)
			return ""
		}
	}
	if iface[len(iface)-1] == '\n' { iface = iface[0:len(iface)-1] }
	return iface
}

func read_dvfs(core uint8, extra_args ...string) uint64 {
	c_str := strconv.Itoa(int(core))
	var out strings.Builder
	var stderr strings.Builder
	cmd := exec.Command("sudo", "rdmsr", "-p " + c_str, "0x199")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { 
		stderr_str := stderr.String()
		fmt.Println(stderr_str)
		//panic(err)
		return 0
	}
	out_str := out.String()
	if out_str[len(out_str)-1] == '\n' { out_str = out_str[0:len(out_str)-1] }
	dvfs_val, err := strconv.ParseInt(out_str, 16, 64)
	if err != nil { panic(err) }
	return uint64(dvfs_val)
}

func write_dvfs(core uint8, val uint64) error {
	c_str := strconv.Itoa(int(core))
	var out strings.Builder
	var stderr strings.Builder
	cmd_str := "sudo wrmsr -p " + c_str + " 0x199 " + strconv.Itoa(int(val))
	cmd := exec.Command("bash", "-c", cmd_str)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	return err
}

func read_rx_usecs(core uint8, iface_args ...string) uint64 {
	var out strings.Builder
	var stderr strings.Builder
	iface := iface_args[0]

	cmd_str := "ethtool -c " + iface + " | grep \"rx-usecs:\" | cut -d ' ' -f2"
	cmd:= exec.Command("bash", "-c", cmd_str)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	out_str := out.String()
	stderr_str := stderr.String()
	if len(out_str) == 0 {
		if len(stderr_str) > 0 {
			fmt.Printf("**** PROBLEM: cannot exec '%v' correctly - ERROR: \n", "ethtool -c enp3s0f0")
			fmt.Println(stderr_str)
			return 0
		}
	}

	if out_str[len(out_str)-1] == '\n' { out_str = out_str[0:len(out_str)-1] }
	rx_usecs_val, err := strconv.Atoi(out_str)
	return uint64(rx_usecs_val)
}

func write_rx_usecs(core uint8, val uint64) error {
	cmd := exec.Command("sudo", "ethtool", "-C", "enp3s0f0", "rx-usecs", strconv.Itoa(int(val))) 
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	return err
}


