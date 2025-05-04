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

func (m *muster) flush_log_files(sheep_id string) {
	sheep := m.pasture[sheep_id]
	label := sheep.label
	index := strconv.Itoa(int(sheep.index))
	log_fname := m.logs_dir + label + "-" + index
	f, err := os.Create(log_fname)
	if err != nil { panic(err) }
	reader := csv.NewReader(f)
	reader.Comma = ' '
	for _, log := range(sheep.logs) {
		sheep.log_f_map[log.id] = f
		sheep.log_reader_map[log.id] = reader
	}
}

func (m *muster) init_log_files(logs_dir string) {
	_, err := os.Stat(logs_dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(logs_dir, 0777)
		if err != nil { panic(err) }
	}
	for _, sheep := range(m.pasture) {
		sheep.log_f_map = make(map[string]*os.File)
		sheep.log_reader_map = make(map[string]*csv.Reader)
		label := sheep.label
		index := strconv.Itoa(int(sheep.index))
		log_fname := logs_dir + label + "-" + index
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


/*
   populates one full memory buffer of log entries
*/
func do_log(shared_log *log, reader *csv.Reader) error {
	var counter uint64
	for counter=0; counter < shared_log.max_size; counter++ {
		var row []string
		var err error
		row, err = reader.Read()
		// here if finished reading file into partial or empty mem_buff
		if err == io.EOF { 
			// return with partially filled mem_buff
			if counter > 0 {
				return nil
			} else {
				return io.EOF
			}
		} else {
			// here if problem reading from log file
			if err != nil { 
				fmt.Println("!!!ERROR!!!", shared_log.id, err)
				return err
			}
		}
		// here if still reading; append log entry read from file
		log_entry := make([]uint64, len(shared_log.metrics))
		for i := range(len(shared_log.metrics)) {
			val, _ := strconv.Atoi(row[i])
			log_entry[i] = uint64(val)
		}
		*shared_log.mem_buff = append(*shared_log.mem_buff, log_entry)
	}
	// here if finished reading one full mem_buff
	return nil
}

func (r_m *remote_muster) sync_with_logger(sheep_id string, log_id string, reader *csv.Reader, logger_func func(*log, *csv.Reader)error, n_iter int) error {
	sheep := r_m.pasture[sheep_id]
	shared_log := sheep.logs[log_id]
	var err error = nil
	for {
		select {
		case <- shared_log.kill_log_chan:
			return nil
		default:		
			err = logger_func(shared_log, reader)
			if err == io.EOF {
				// do nothing if nothing has been logged yet
				time.Sleep(time.Second)
			} else {
				if err != nil {
					fmt.Println("HERE UNKNOWN SYNC ERROR", shared_log.id)
					return err
				}
				r_m.full_buff_chan <- []string{sheep.id, shared_log.id}
				<- shared_log.ready_buff_chan
				*(shared_log.mem_buff) = make([][]uint64, 0)
			}
		}
	}
}

func (r_m *remote_muster) log_manage(sheep_id string, log_id string, cmd string, logger_id string) bool {
	sheep := r_m.pasture[sheep_id]
	log := sheep.logs[log_id]
	logger_func := r_m.native_loggers[logger_id]
	logs_dir := r_m.logs_dir

	switch {
	case cmd == "start":
		// start communication with native logger
		go logger_func(sheep, log, logs_dir)
		log.ready_request_chan <- true
		return true

	case cmd == "stop":
		// stop communication with native logger
		log.kill_log_chan <- true
		sheep.detach_native_logger <- true
		log.ready_request_chan <- true
		return true

	case cmd == "close":
		label := sheep.label
		index := strconv.Itoa(int(sheep.index))
		log_fname := logs_dir + label + "-" + index
		f, err := os.Create(log_fname)
		if err != nil { panic(err) }
		reader := csv.NewReader(f)
		reader.Comma = ' '
		for _, log := range(sheep.logs) {
			sheep.log_f_map[log.id] = f
			sheep.log_reader_map[log.id] = reader
		}

		log.ready_request_chan <- true
		return true

	case cmd == "all":
		f := sheep.log_f_map[log_id]
		reader := sheep.log_reader_map[log_id]
		f.Seek(0, io.SeekStart)

		// at runtime, there are as many of this thread as there are sheep x num_logs_per_sheep
		go func() {
			sheep := sheep
			log := log
			reader := reader
			err := r_m.sync_with_logger(sheep.id, log.id, reader, do_log, -1)
			if err != nil { 
				panic(err) 
			} else { 
				// assume log was killed 
				fmt.Println("LOG WAS KILLED: ", sheep.id, log.id)
				return
			}
		} ()

		log.ready_request_chan <- true 
		return true

//	case cmd == "first":
//		// get first instance from native logger
//		go func() {
//			sheep_id := sheep_id
//			log_id := log_id
//			f := r_m.pasture[sheep_id].log_f_map[log_id]
//			reader := r_m.pasture[sheep_id].log_reader_map[log_id]
//			f.Seek(0, io.SeekStart)
//			err := r_m.sync_with_logger(sheep_id, log_id, reader, do_log, 1)
//			if err == io.EOF {
//				fmt.Println("************** FILE IS EMPTY *************", log_id) 
//				r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
//				return
//			}
//			if err != nil { panic(err) }
//			r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
//		} ()
//	case cmd == "last":
//		// get last instance from native logger
//		go func() {
//			sheep_id := sheep_id
//			log_id := log_id
//			f := r_m.pasture[sheep_id].log_f_map[log_id]
//			f.Seek(0, io.SeekStart)
//			// get length of log file
//			reader1 := csv.NewReader(f)
//			reader1.Comma = ' '
//			rows, err := reader1.ReadAll() 
//			if err != nil { panic(err) }
//			len_rows := len(rows)
//			if len_rows == 0 {
//				fmt.Println("************** FILE IS EMPTY *************", log_id) 
//				r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
//				return
//			}
//
//			// read log file except last entry
//			f.Seek(0, io.SeekStart)
//			reader2 := csv.NewReader(f)
//			reader2.Comma = ' '
//			counter := 0
//			for {
//				if counter == len_rows - 1 { break }
//				_, err := reader2.Read()
//				if err == io.EOF { 
//					fmt.Println("************** FILE IS EMPTY *************", log_id) 
//					break
//				}
//				if err != nil { panic(err) }
//				counter ++
//			}
//
//			err = r_m.sync_with_logger(sheep_id, log_id, reader2, do_log, 1)
//			if err != nil { panic(err) }
//			r_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
//		} ()

	default:
		fmt.Println("************ UNKNOWN LOG COMMAND: ", cmd)
		return false
	}
}


/****/

func (r_m *remote_muster) ctrl_manage(sheep_id string) {
	fmt.Printf("\033[36m-- MUSTER %v -- SHEEP %v - STARTING CONTROL MANAGER\n\033[0m", r_m.id, sheep_id)
	sheep := r_m.pasture[sheep_id]
	var err error
	for {
		select {
		case new_ctrls := <- sheep.new_ctrl_chan:
			for ctrl_id, ctrl_val := range(new_ctrls) {
				fmt.Println(sheep)
				fmt.Println(sheep.controls)
				fmt.Println(ctrl_id)
				fmt.Println(sheep.controls[ctrl_id])
				fmt.Println(sheep.controls[ctrl_id].setter)
				err = sheep.controls[ctrl_id].setter(sheep.core, ctrl_val)
				if err != nil { panic(err) }
				sheep.controls[ctrl_id].value = ctrl_val
			}
			sheep.done_ctrl_chan <- control_reply{done: true, ctrls: new_ctrls}
		}
	}
}









