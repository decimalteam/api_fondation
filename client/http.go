package client

import (
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
)

func GetRequest(requestURI string) []byte {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(requestURI)
	// fasthttp does not automatically request a gzipped response.
	// We must explicitly ask for it.
	req.Header.Set("Accept-Encoding", "gzip")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the request
	err := fasthttp.Do(req, resp)
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		return nil
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		fmt.Printf("Expected status code %d but got %d\n", fasthttp.StatusOK, resp.StatusCode())
		return nil
	}

	// Verify the content type
	contentType := resp.Header.Peek("Content-Type")
	if bytes.Index(contentType, []byte("application/json")) != 0 {
		fmt.Printf("Expected content type application/json but got %s\n", contentType)
		return nil
	}

	// Do we need to decompress the response?
	contentEncoding := resp.Header.Peek("Content-Encoding")
	var res []byte
	if bytes.EqualFold(contentEncoding, []byte("gzip")) {
		res, err = resp.BodyGunzip()
		if err != nil {
			fmt.Printf("gunzip response body failed: %s\n", err)
		}
	} else {
		res = resp.Body()
	}

	return res
}
