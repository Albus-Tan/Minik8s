package httpclient

import (
	"encoding/json"
	"io"
	"log"
)

func ReadAndUnmarshal(body io.ReadCloser, target interface{}) error {
	content, err := io.ReadAll(body)
	defer func(body io.ReadCloser) {
		err = body.Close()
		if err != nil {
			log.Printf("[utils][http][ReadAndUnmarshal] http response body.Close() failed, %v\n", err)
		}
	}(body)
	if err != nil {
		log.Printf("[utils][http][ReadAndUnmarshal] reading http response failed, %v\n", err)
		return err
	}
	err = json.Unmarshal(content, target)
	if err != nil {
		log.Printf("[utils][http][ReadAndUnmarshal] http response json.Unmarshal failed, target type mismatch: %v\n", err)
		return err
	}
	return nil
}
