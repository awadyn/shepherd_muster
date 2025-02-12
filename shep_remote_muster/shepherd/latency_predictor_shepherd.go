package main

/**************************************/

type latency_predictor_shepherd struct {
	intlog_shepherd 
	lat_pred_musters map[string]*latency_predictor_muster 
}

func (lat_pred_s *latency_predictor_shepherd) init() {
	lat_pred_s.lat_pred_musters = make(map[string]*latency_predictor_muster)
	for _, intlog_m := range(lat_pred_s.intlog_musters) {
		lat_pred_m := latency_predictor_muster{intlog_muster: *intlog_m}
		lat_pred_m.init()
		lat_pred_s.lat_pred_musters[lat_pred_m.id] = &lat_pred_m
		lat_pred_m.show()
	}
}


