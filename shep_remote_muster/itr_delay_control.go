package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"bytes"
)

/*********************************************/
func get_internal_iface() string {
	var out strings.Builder
	var stderr strings.Builder
	// TODO need to unify this across different interface naming conventions
//	cmd:= exec.Command("bash", "-c", "ls /sys/class/net | grep enp | grep f0")
	cmd:= exec.Command("bash", "-c", "ip addr | grep 'state UP' | cut -d ':' -f2 | cut -d ' ' -f2")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	iface := out.String()
	stderr_str := stderr.String()
	if len(iface) == 0 {
		if len(stderr_str) > 0 {
			fmt.Printf("**** PROBLEM: cannot read ethernet interface name.. aborting..\n")
			return ""
		}
	}
	if iface[len(iface)-1] == '\n' { iface = iface[0:len(iface)-1] }
	return iface
}


func read_rx_usecs(core uint8, iface_args ...string) uint64 {
	var out strings.Builder
	var stderr strings.Builder
	iface := get_internal_iface()
	if iface == "" {
		fmt.Printf("**** PROBLEM: cannot get internal network interface name.. aborting\n")
		return 0
	}

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

func write_rx_usecs(core uint8, val uint64, iface_args ...string) error {
	iface := get_internal_iface()
	if iface == "" {
		fmt.Printf("**** PROBLEM: cannot get internal network interface name.. aborting\n")
		return nil
	}

	cmd_str := "sudo ethtool -C " + iface + " rx-usecs " + strconv.Itoa(int(val))
	cmd:= exec.Command("bash", "-c", cmd_str)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	return err
}


