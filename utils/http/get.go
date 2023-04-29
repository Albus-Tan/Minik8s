package httpclient

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func GetAndUnmarshal(URL string, target interface{}) error {
	resp, err := http.Get(URL)
	if err != nil {
		log.Println("[utils][http][GetAndUnmarshal] http.Get failed", err)
		return err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[utils][http][GetAndUnmarshal] http response io.ReadAll(resp.Body) failed", err)
		return err
	}

	err = json.Unmarshal(content, target)
	if err != nil {
		log.Println("[utils][http][GetAndUnmarshal] http response json.Unmarshal failed", err)
		return err
	}

	return nil
}
