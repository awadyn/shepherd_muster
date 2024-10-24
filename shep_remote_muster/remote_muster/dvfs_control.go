package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

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


