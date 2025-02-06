package main

/**************************************/

type latency_predictor_muster struct {
	intlog_muster
}

type latency_predictor_shepherd struct {
	intlog_shepherd 
}

func (lat_pred_s *latency_predictor_shepherd) init() {
	for _, intlog_m := range(lat_pred_s.intlog_musters) {
		lat_pred_m := latency_predictor_muster{intlog_muster: *intlog_m}
		lat_pred_m.init()
		lat_pred_m.show()
	}
}


