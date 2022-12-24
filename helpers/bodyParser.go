package helpers

import (
	"encoding/json"
	"reflect"
	"sync"
)

func BodyParser(body interface{}, queryType string) []interface{} {
	var mutex sync.RWMutex
	convertToMapping := make(map[string]interface{})

	convertToJson, _ := json.Marshal(body)
	json.Unmarshal(convertToJson, &convertToMapping)

	storeRequest := []interface{}{}

	if queryType == "insert" {
		for _, v := range convertToMapping {
			if v != "0001-01-01T00:00:00Z" && !reflect.DeepEqual(v, float64(0)) && !reflect.DeepEqual(v, nil) || (reflect.DeepEqual(v, reflect.Struct) && !reflect.DeepEqual(reflect.TypeOf(0).NumField(), float64(0))) {
				mutex.Lock()
				storeRequest = append(storeRequest, v)
				mutex.Unlock()
			}
		}
	}

	if queryType == "update" {
		id := convertToMapping["id"]
		delete(convertToMapping, "id")

		for _, v := range convertToMapping {
			if v != "0001-01-01T00:00:00Z" && !reflect.DeepEqual(v, float64(0)) && !reflect.DeepEqual(v, nil) || (reflect.DeepEqual(v, reflect.Struct) && !reflect.DeepEqual(reflect.TypeOf(0).NumField(), float64(0))) {
				mutex.Lock()
				storeRequest = append(storeRequest, v)
				mutex.Unlock()
			}
		}

		if !reflect.DeepEqual(id, nil) {
			storeRequest = append(storeRequest, id)
		}
	}

	return storeRequest
}
