package api

import (
	"encoding/json"
	"io"
	"log"
	"minik8s/pkg/api/types"
	"net/http"
)

type IResponse interface {
	FillResponse(resp *http.Response) error
}

type Response struct {
	Status   string `json:"status,omitempty"`
	ErrorMsg string `json:"error,omitempty"`
}

func (r *Response) FillResponse(resp *http.Response) error {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("FillResponse io.ReadAll(resp.Body) failed", err)
		return err
	}

	err = json.Unmarshal(buf, r)
	if err != nil {
		log.Println("FillResponse json.Unmarshal failed", err)
		return err
	}

	return nil
}

type PostResponse struct {
	Response        `json:",inline"`
	UID             types.UID `json:"uid,omitempty"`
	ResourceVersion string    `json:"resourceVersion,omitempty"`
}

func (r *PostResponse) FillResponse(resp *http.Response) error {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("FillResponse io.ReadAll(resp.Body) failed", err)
		return err
	}

	err = json.Unmarshal(buf, r)
	if err != nil {
		log.Println("FillResponse json.Unmarshal failed", err)
		return err
	}

	return nil
}

type PutResponse struct {
	Response        `json:",inline"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

func (r *PutResponse) FillResponse(resp *http.Response) error {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("FillResponse io.ReadAll(resp.Body) failed", err)
		return err
	}

	err = json.Unmarshal(buf, r)
	if err != nil {
		log.Println("FillResponse json.Unmarshal failed", err)
		return err
	}

	return nil
}

type DeleteResponse struct {
	Response `json:",inline"`
}
