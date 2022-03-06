package core

func UpdateInMask(update map[string]interface{}, mask []string) (badField string, ok bool) {
updateLoop:
	for k := range update {
		for _, v := range mask {
			if k == v {
				continue updateLoop
			}
		}
		return k, false
	}
	return "", true
}

func updateApplyRemap(update map[string]interface{}, remap map[string]string) {
	for from, to := range remap {
		v, ok := update[from]
		if !ok {
			continue
		}
		update[to] = v
		delete(update, from)
	}
}
