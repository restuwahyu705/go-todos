package helpers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
)

func ScanStructMutation(r *http.Request, body interface{}) (cols []string, args []interface{}, err error) {
	var mutex sync.RWMutex
	readerKey := []string{}
	readerValue := []interface{}{}
	structKey := []interface{}{}
	countKeyMatch := 0

	readerToMap := make(map[string]interface{})
	structToMap := make(map[string]interface{})

	reader, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	if err := json.Unmarshal(reader, &readerToMap); err != nil {
		return nil, nil, err
	}

	for k, v := range readerToMap {
		mutex.RLock()
		readerKey = append(readerKey, k)
		readerValue = append(readerValue, v)
		mutex.RUnlock()
	}

	toStringify, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}

	if err := json.Unmarshal(toStringify, &structToMap); err != nil {
		return nil, nil, err
	}

	for k, _ := range structToMap {
		mutex.RLock()
		structKey = append(structKey, k)
		mutex.RUnlock()
	}

	for _, sv := range structKey {
		for _, rv := range readerKey {
			if sv == rv {
				mutex.RLock()
				countKeyMatch++
				mutex.RUnlock()
				break
			}
		}
	}

	if countKeyMatch != len(readerKey) {
		return nil, nil, errors.New("a number of request key not match with struct key")
	}

	return readerKey, readerValue, nil
}
