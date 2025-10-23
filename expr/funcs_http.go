package expr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Second*2)
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
)

var httpRequest = FuncDefine5WithCtx(func(c *Context, method string, url string, headers map[string]any, body any, timeoutMillSec float64) map[string]any {

	var bb []byte
	switch bd := body.(type) {
	case []byte:
		bb = bd
	case string:
		bb = ToBytes(bd)
	case nil:
	default:
		bb, _ = json.Marshal(bd)
	}
	if timeoutMillSec <= 0 {
		timeoutMillSec = 30000
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(timeoutMillSec)*time.Millisecond)
	defer cancel()
	res := map[string]any{}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bb))
	if err != nil {
		res["err"] = err.Error()
		return res
	}
	for k, v := range headers {
		req.Header.Set(k, StringOf(v))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		res["err"] = err.Error()
		return res
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		res["err"] = err.Error()
		return res
	}
	res["body"] = bs
	hds := map[string]any{}
	for key, val := range resp.Header {
		if len(val) > 0 {
			hds[key] = val[0]
		}
	}
	//if resp.Header.Get("Content-Type") == "application/json" {
	//	var i any
	//	err = json.Unmarshal(bs, &i)
	//	if err != nil {
	//		res["err"] = "invalid json response from http request"
	//	}
	//	res["json"] = i
	//}
	res["header"] = hds
	res["status"] = float64(resp.StatusCode)
	return res
})
