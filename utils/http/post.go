package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

func PostJson(URL string, content interface{}) (*http.Response, error) {
	cli := &http.Client{}
	b, err := json.Marshal(content)
	if err != nil {
		log.Println("[utils][http][PostJson] json.Marshal failed", err)
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(b))
	if err != nil {
		log.Println("[utils][http][PostJson] http.NewRequest create failed", err)
		return nil, err
	}
	return cli.Do(req)
}

func PostString(URL string, content string) (*http.Response, error) {
	cli := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader([]byte(content)))
	if err != nil {
		log.Println("[utils][http][PostString] http.NewRequest create failed", err)
		return nil, err
	}
	return cli.Do(req)
}

func PostForm(URL string, form map[string]string) string {
	values := url.Values{}
	for key, value := range form {
		values.Add(key, value)
	}

	var err error
	var resp *http.Response
	if resp, err = http.PostForm(URL, values); err == nil {
		defer resp.Body.Close()
		var body []byte
		if body, err = io.ReadAll(resp.Body); err == nil {
			return string(body)
		}
	}
	return err.Error()
}
