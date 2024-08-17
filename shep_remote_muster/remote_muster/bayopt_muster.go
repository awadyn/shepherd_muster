package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"encoding/csv"
	"time"
	"io"
	"strings"
	"bytes"
//	"slices"
//	"bufio"
)

/*********************************************/


func (bayopt_m *bayopt_muster) init() {
	bayopt_m.intlog_metrics = []string{"i", "rx_desc", "rx_bytes", "tx_desc", "tx_bytes",
				    "instructions", "cycles", "ref_cycles", "llc_miss", 
				    "c1", "c1e", "c3", "c3e", "c6", "c7", "joules","timestamp"}
	bayopt_m.buff_max_size = 1
	var core uint8
	var rx_usecs_shared uint64

	// get name of internal network interface
	iface := bayopt_m.get_internal_iface()
	if iface == "" {
		fmt.Printf("**** PROBLEM: %v cannot get internal network interface name.. aborting\n", bayopt_m.id)
		return
	}

	ctrl_itr_shared := control{n_ip: bayopt_m.ip, knob: "itr-delay", dirty: false}  
	rx_usecs_reading := ctrl_itr_shared.getter(0, read_rx_usecs, iface)  

	if rx_usecs_reading == 0 {
		fmt.Printf("**** PROBLEM: %v cannot read ethtool rx_usecs value.. assuming default value 1\n", bayopt_m.id)
		rx_usecs_shared = 1
	} else {
		rx_usecs_shared = rx_usecs_reading
	}
	for core = 0; core < bayopt_m.ncores; core ++ {
		mem_buff := make([][]uint64, bayopt_m.buff_max_size)
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + bayopt_m.ip

		log_id := "log-" + c_str + "-" + bayopt_m.ip
		log_c := log{id: log_id,
			     metrics: bayopt_m.intlog_metrics,
			     //metrics: bayopt_m.bayopt_metrics,
			     max_size: bayopt_m.buff_max_size,
			     mem_buff: &mem_buff,
			     kill_log_chan: make(chan bool, 1),
			     request_log_chan: make(chan string),
			     done_log_chan: make(chan bool, 1),
			     ready_buff_chan: make(chan bool, 1)}

		ctrl_dvfs := control{n_ip: bayopt_m.ip, knob: "dvfs", dirty: false}
		ctrl_dvfs.id = "ctrl-dvfs-" + c_str + "-" + bayopt_m.ip
		ctrl_dvfs.value = ctrl_dvfs.getter(core, read_dvfs)
		ctrl_itr := ctrl_itr_shared
		ctrl_itr.id = "ctrl-itr-" + c_str + "-" + bayopt_m.ip
		ctrl_itr.value = rx_usecs_shared

		bayopt_m.pasture[sheep_id].logs[log_id] = &log_c
		bayopt_m.pasture[sheep_id].controls[ctrl_dvfs.id] = &ctrl_dvfs
		bayopt_m.pasture[sheep_id].controls[ctrl_itr.id] = &ctrl_itr
	}

	bayopt_m.init_log_files()
}

func (bayopt_m *bayopt_muster) init_log_files() {
	bayopt_m.log_f_map = make(map[string]*os.File)
	bayopt_m.log_reader_map = make(map[string]*csv.Reader)
	for sheep_id, _ := range(bayopt_m.pasture) {
		core := bayopt_m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := bayopt_m.logs_dir + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		bayopt_m.log_f_map[sheep_id] = f
	}
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
		panic(err) 
	}
	out_str := out.String()
	if out_str[len(out_str)-1] == '\n' { out_str = out_str[0:len(out_str)-1] }
	dvfs_val, err := strconv.ParseInt(out_str, 16, 64)
	if err != nil { panic(err) }
	return uint64(dvfs_val)
}

func write_dvfs(core uint8, val uint64) error {
	c_str := strconv.Itoa(int(core))
	cmd := exec.Command("sudo", "wrmsr", "-p " + c_str, "0x199", strconv.Itoa(int(val)))
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

func write_rx_usecs(core uint8, val uint64) error {
	cmd := exec.Command("sudo", "ethtool", "-C", "enp3s0f0", "rx-usecs", strconv.Itoa(int(val))) 
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil { panic(err) }
	return err
}

func do_bayopt_log(shared_log *log, reader *csv.Reader) error {
	*shared_log.mem_buff = make([][]uint64, 0)
	var counter uint64 = 0
	for {
		switch {
		case counter < shared_log.max_size:
			var row []string
			var err error
			for {
				row, err = reader.Read()
				if err == io.EOF { 
					time.Sleep(time.Second * 2)
					continue
				}
				if err != nil { panic(err) }
				break
			}
			*shared_log.mem_buff = append(*shared_log.mem_buff, []uint64{})
			for i := range(len(shared_log.metrics)) {
				val, _ := strconv.Atoi(row[i])
				(*shared_log.mem_buff)[counter] = append((*shared_log.mem_buff)[counter], uint64(val))
			}
			counter ++
		case counter == shared_log.max_size:
			return nil
		}
	}
}

func (bayopt_m *bayopt_muster) assign_log_files(sheep_id string) {
	c_str := strconv.Itoa(int(bayopt_m.pasture[sheep_id].core))
	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str
	if bayopt_m.log_f_map[sheep_id] != nil {
		bayopt_m.log_f_map[sheep_id].Close()
	}
	f, err := os.Create(log_fname)
	if err != nil { panic(err) }
	bayopt_m.log_f_map[sheep_id] = f
}


/* The following functions associate each sheep with its
   intlog data in /proc/ixgbe_stats/core/sheep.core
*/
func (bayopt_m *bayopt_muster) attach_native_logger(sheep_id string) {
	c_str := strconv.Itoa(int(bayopt_m.pasture[sheep_id].core))
	src_fname := "/proc/ixgbe_stats/core/" + c_str
	log_fname := "/users/awadyn/shepherd_muster/shep_remote_muster/intlog_logs/" + c_str

	cmd := exec.Command("bash", "-c", "cat " + src_fname)
	if err := cmd.Run(); err != nil { panic(err) }
	for {
		select {
		case <- bayopt_m.pasture[sheep_id].detach_native_logger:
			return
		default:
			cmd = exec.Command("bash", "-c", "cat " + src_fname + " >> " + log_fname)
			if err := cmd.Run(); err != nil { panic(err) }
			time.Sleep(time.Second * 2)
		}
	}
}


func (bayopt_m *bayopt_muster) ctrl_manage(sheep_id string) {
	fmt.Printf("-- -- STARTING CONTROL MANAGER FOR SHEEP %v\n", sheep_id)
	sheep := bayopt_m.pasture[sheep_id]
	var err error
	for {
		select {
		case new_ctrls := <- sheep.request_ctrl_chan:
			for ctrl_id, ctrl_val := range(new_ctrls) {
				switch {
				case sheep.controls[ctrl_id].knob == "dvfs":
					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val, write_dvfs)
				case sheep.controls[ctrl_id].knob == "itr-delay":
					err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val, write_rx_usecs)
				default:
				}
				if err != nil { panic(err) }
				sheep.controls[ctrl_id].value = ctrl_val
			}
			bayopt_m.pasture[sheep.id].ready_ctrl_chan <- true
		}
	}
}


func (bayopt_m *bayopt_muster) log_manage(sheep_id string, log_id string) {
	fmt.Printf("-- -- STARTING LOG MANAGER FOR SHEEP %v - LOG %v \n", sheep_id, log_id)
	for {
		select {
		case cmd := <- bayopt_m.pasture[sheep_id].logs[log_id].request_log_chan:
			switch {
			case cmd == "start":
				// start communication with native logger
				bayopt_m.assign_log_files(sheep_id)
				go bayopt_m.attach_native_logger(sheep_id)
				bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
			case cmd == "stop":
				// stop communication with native logger
				bayopt_m.pasture[sheep_id].detach_native_logger <- true
				bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := bayopt_m.log_f_map[sheep_id]
					reader := csv.NewReader(f)
					reader.Comma = ' '
					f.Seek(0, io.SeekStart)
					err := bayopt_m.sync_with_logger(sheep_id, log_id, reader, do_bayopt_log, 1)
					if err == io.EOF {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
						return
					}
					if err != nil { panic(err) }
					bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := bayopt_m.log_f_map[sheep_id]
					f.Seek(0, io.SeekStart)

					// get length of log file
					reader1 := csv.NewReader(f)
					reader1.Comma = ' '
					rows, err := reader1.ReadAll() 
					if err != nil { panic(err) }
					len_rows := len(rows)
					if len_rows == 0 {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
						return
					}

					// read log file except last entry
					f.Seek(0, io.SeekStart)
					reader2 := csv.NewReader(f)
					reader2.Comma = ' '
					counter := 0
					for {
						if counter == len_rows - 1 { break }
						_, err := reader2.Read()
						if err == io.EOF { 
							fmt.Println("************** FILE IS EMPTY *************", log_id) 
							break
						}
						if err != nil { panic(err) }
						counter ++
					}

					err = bayopt_m.sync_with_logger(sheep_id, log_id, reader2, do_bayopt_log, 1)
					if err != nil { panic(err) }
					bayopt_m.pasture[sheep_id].logs[log_id].done_log_chan <- true
				} ()
			default:
				fmt.Println("************ UNKNOWN BAYOPT_LOG COMMAND: ", cmd)
			}
		}
	}
}






