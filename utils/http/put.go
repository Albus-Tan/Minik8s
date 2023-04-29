package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func PutForm(URL string, form map[string]string) (string, error) {
	cli := &http.Client{}
	formJson, err := json.Marshal(form)
	if err != nil {
		log.Println("[utils][http][PutForm] form json.Marshal failed", err)
		return "", err
	}

	r := bytes.NewReader(formJson)
	req, err := http.NewRequest(http.MethodPut, URL, r)
	if err != nil {
		log.Println("[utils][http][PutForm] form http.NewRequest create failed", err)
		return "", err
	}

	resp, err := cli.Do(req)
	if err != nil {
		log.Println("[utils][http][PutForm] form http request send failed", err)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[utils][http][PutForm] form http request read response body failed", err)
		return "", err
	}

	return string(body), nil
}

func PutJson(URL string, v interface{}) (*http.Response, error) {
	cli := &http.Client{}
	vJson, err := json.Marshal(v)
	if err != nil {
		log.Println("[utils][http][PutJson] http body json.Marshal failed", err)
		return nil, err
	}

	r := bytes.NewReader(vJson)
	req, err := http.NewRequest(http.MethodPut, URL, r)
	if err != nil {
		log.Println("[utils][http][PutJson] http.NewRequest failed", err)
		return nil, err
	}

	return cli.Do(req)
}
