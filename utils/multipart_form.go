package utils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
)

func NewMultipartForm(uri string, params map[string]string, buf *[]byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	f, _ := writer.CreateFormField("media")
	_, e := io.Copy(f, bytes.NewReader(*buf))
	if e != nil {
		return nil, e
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	_ = writer.Close()

	req, _ := http.NewRequest("POST", uri, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	return req, nil
}
