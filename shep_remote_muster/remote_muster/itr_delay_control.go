package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"bytes"
)

/*********************************************/


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


