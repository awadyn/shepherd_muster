package main

/*********************************************/
func (stats_m *stats_muster) init() {
	stats_m.rx_bytes_all = make(map[string][]uint64)
	stats_m.processing_lock = make(chan bool, 1)
	stats_m.processing_lock <- true
}

