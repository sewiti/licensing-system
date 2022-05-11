package core

import "encoding/json"

func ChangesInMask(changes map[string]struct{}, mask []string) (badField string, ok bool) {
changesLoop:
	for k := range changes {
		for _, v := range mask {
			if k == v {
				continue changesLoop
			}
		}
		return k, false
	}
	return "", true
}

func UnmarshalChanges(data []byte) (map[string]struct{}, error) {
	v := make(map[string]interface{})
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	changes := make(map[string]struct{}, len(v))
	for k := range v {
		changes[k] = struct{}{}
	}
	return changes, nil
}
