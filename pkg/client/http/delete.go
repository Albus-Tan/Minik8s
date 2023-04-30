package httpclient

import (
	"io"
	"log"
	"net/http"
)

func Delete(URL string) (string, error) {
	cli := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, URL, nil)
	if err != nil {
		log.Println("[utils][http][Delete] http.NewRequest create failed", err)
		return "", err
	}

	resp, err := cli.Do(req)
	if err != nil {
		log.Println("[utils][http][Delete] http request send failed", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[utils][http][Delete] http request read response body failed", err)
		return "", err
	}

	return string(body), nil
}
