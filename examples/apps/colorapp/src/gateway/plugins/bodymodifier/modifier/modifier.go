// Package modifier exposes a request modifier for generating bodies
// from the querystring params
package modifier

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/martian/parse"
	kazaam "gopkg.in/qntfy/kazaam.v3"
)

type Config struct {
	Template    string               `json:"template"`
	Method      string               `json:"method"`
	Scope       []parse.ModifierType `json:"scope"`
	ContentType string               `json:"content_type"`
}

type BodyModifier struct {
	template    string
	method      string
	contentType string
	Scope       []parse.ModifierType `json:"scope"`
}

func (m *BodyModifier) ModifyRequest(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	payloadBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body.Close()
	k, err := kazaam.NewKazaam(m.template)
	if err != nil {
		return err
	}

	tranformedDataBytes, err := k.TransformJSONString(string(payloadBytes))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.Read(tranformedDataBytes)
	if m.method != "" {
		req.Method = m.method
	}
	if m.contentType != "" {
		req.Header.Set("Content-Type", m.contentType)
	} else {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	req.ContentLength = int64(len(tranformedDataBytes))
	req.Body = ioutil.NopCloser(bytes.NewReader(tranformedDataBytes))

	return nil
}

func (m *BodyModifier) ModifyResponse(res *http.Response) error {
	if res.Body == nil {
		return nil
	}

	responseBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	k, err := kazaam.NewKazaam(m.template)
	if err != nil {
		return err
	}

	tranformedDataBytes, err := k.TransformJSONString(string(responseBytes))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.Read(tranformedDataBytes)
	if m.contentType != "" {
		res.Header.Set("Content-Type", m.contentType)
	} else {
		res.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	res.ContentLength = int64(len(tranformedDataBytes))
	res.Body = ioutil.NopCloser(bytes.NewReader(tranformedDataBytes))

	return nil
}

func FromJSON(b []byte) (*BodyModifier, error) {
	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	bytes, err := base64.StdEncoding.DecodeString(cfg.Template) // Converting data
	if err != nil {
		return nil, err
	}
	log.Println("FromJSON", cfg)
	return &BodyModifier{
		template:    string(bytes),
		method:      cfg.Method,
		Scope:       cfg.Scope,
		contentType: cfg.ContentType,
	}, nil
}
