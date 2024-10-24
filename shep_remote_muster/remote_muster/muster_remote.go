package main

import (
	"fmt"
	"time"
	"io"
	"strconv"
	"encoding/csv"
	"os"
)

/*********************************************/

func (m *muster) init_remote() {
	for _, sheep := range(m.pasture) {
		sheep.new_ctrl_chan = make(chan map[string]uint64)
		sheep.request_log_chan = make(chan []string)
		sheep.request_ctrl_chan = make(chan string)
		sheep.detach_native_logger = make(chan bool, 1)
		//sheep.done_kill_chan = make(chan bool, 1)
	}
}

func (m *muster) init_log_files(logs_dir string) {
	err := os.Mkdir(logs_dir, 0750)
	if err != nil && !os.IsExist(err) { panic(err) }
	for sheep_id, sheep := range(m.pasture) {
		sheep.log_f_map = make(map[string]*os.File)
		sheep.log_reader_map = make(map[string]*csv.Reader)
		core := m.pasture[sheep_id].core
		c_str := strconv.Itoa(int(core))
		log_fname := logs_dir + c_str
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		reader := csv.NewReader(f)
		reader.Comma = ' '
		for _, log := range(sheep.logs) {
			sheep.log_f_map[log.id] = f
			sheep.log_reader_map[log.id] = reader
		}
	}
}

func (m *muster) cleanup() {
	for _, sheep := range(m.pasture) {
		for _, f := range(sheep.log_f_map) {
			f.Close()
		}
	}
}

//func (ctrl *control) getter(core uint8, get_func func(uint8, ...string)uint64, args ...string) uint64 {
//	var ctrl_val uint64
//	if len(args) > 0 {
//		ctrl_val = get_func(core, args[0])
//	} else {
//		ctrl_val = get_func(core)
//	}
//	return ctrl_val
//}
//func (ctrl *control) setter(core uint8, value uint64, set_func func(uint8, uint64)error) error {
//	err := set_func(core, value)
//	return err
//}

/*
   populates one full memory buffer of log entries
*/
func do_log(shared_log *log, reader *csv.Reader) error {
	*shared_log.mem_buff = make([][]uint64, 0)
	var counter uint64 = 0
	for {
		switch {
		case counter < shared_log.max_size:
			var row []string
			var err error
			row, err = reader.Read()
			if err == io.EOF { 
				return err
			}
			if err != nil { panic(err) }
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

func (m *muster) sync_with_logger(sheep_id string, log_id string, reader *csv.Reader, logger_func func(*log, *csv.Reader)error, n_iter int) error {
	shared_log := m.pasture[sheep_id].logs[log_id] 
	var err error = nil
	iter := 0
	for {
		switch {
		case n_iter > 0:
			if iter == n_iter { return err }
			iter ++
		default:
		}
		err = logger_func(shared_log, reader)
		switch {
		case err == nil:	// => logged one buff
			m.full_buff_chan <- []string{sheep_id, log_id}
			<- shared_log.ready_buff_chan
			*shared_log.mem_buff = make([][]uint64, shared_log.max_size)
		case err == io.EOF:	// => log reader at EOF
			// do nothing if nothing has been logged yet
			if len(*(shared_log.mem_buff)) == 0 { return io.EOF }
			// otherwise sync whatever has been logged with mirror
			m.full_buff_chan <- []string{sheep_id, log_id}
			<- shared_log.ready_buff_chan
			return io.EOF
		}
	}
}

func (r_m *remote_muster) log_manage(sheep_id string, logs_dir string, native_log func(*sheep, *log, string)) {
	fmt.Printf("\033[34m-- MUSTER %v --SHEEP %v - STARTING LOG MANAGER\n\033[0m", r_m.id, sheep_id)
	var err error
	for {
		select {
		case req := <- r_m.pasture[sheep_id].request_log_chan:
			log_id := req[0]
			cmd := req[1]
			sheep := r_m.pasture[sheep_id]
			log := sheep.logs[log_id]
			switch {
			case cmd == "start":
				// start communication with native logger
				go native_log(sheep, log, logs_dir)
				if err != nil { 
					r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- false
				} else {
					r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				}
			case cmd == "stop":
				select {
				case r_m.pasture[sheep_id].logs[log_id].kill_log_chan <- true:
				default:
				}
				// stop communication with native logger
				r_m.pasture[sheep_id].detach_native_logger <- true
				r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
			case cmd == "first":
				// get first instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := r_m.pasture[sheep_id].log_f_map[log_id]
					reader := r_m.pasture[sheep_id].log_reader_map[log_id]
					f.Seek(0, io.SeekStart)
					err := r_m.sync_with_logger(sheep_id, log_id, reader, do_log, 1)
					if err == io.EOF {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
						return
					}
					if err != nil { panic(err) }
					r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			case cmd == "last":
				// get last instance from native logger
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := r_m.pasture[sheep_id].log_f_map[log_id]
					f.Seek(0, io.SeekStart)
					// get length of log file
					reader1 := csv.NewReader(f)
					reader1.Comma = ' '
					rows, err := reader1.ReadAll() 
					if err != nil { panic(err) }
					len_rows := len(rows)
					if len_rows == 0 {
						fmt.Println("************** FILE IS EMPTY *************", log_id) 
						r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
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

					err = r_m.sync_with_logger(sheep_id, log_id, reader2, do_log, 1)
					if err != nil { panic(err) }
					r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
				} ()
			case cmd == "all":
				go func() {
					sheep_id := sheep_id
					log_id := log_id
					f := r_m.pasture[sheep_id].log_f_map[log_id]
					reader := r_m.pasture[sheep_id].log_reader_map[log_id]
					f.Seek(0, io.SeekStart)
					//r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
					var first_iter bool = true
					var kill_log bool = false
					for {
						select {
						case <- r_m.pasture[sheep_id].logs[log_id].kill_log_chan:
							kill_log = true
							continue
						default:
							err := r_m.sync_with_logger(sheep_id, log_id, reader, do_log, -1)
							if err == io.EOF {
								time.Sleep(time.Second * 2)
							} else { 
								if err != nil { panic(err) } 
							}
							if first_iter { r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true }

						}
						if kill_log { 
							r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
							return 
						}
					}
				} ()
			default:
				fmt.Println("************ UNKNOWN LOG COMMAND: ", cmd)
			}
		}
	}
}

/****/

func (m *muster) sync_new_ctrl() {
	for {
		select {
		case new_ctrl_req := <- m.new_ctrl_chan:
			sheep := m.pasture[new_ctrl_req.sheep_id]
			new_ctrls := new_ctrl_req.ctrls
			sheep.new_ctrl_chan <- new_ctrls
		}
	}
}








