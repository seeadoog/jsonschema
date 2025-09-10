package expr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

var httpRequest = FuncDefine5(func(method string, url string, headers map[string]any, body any, timeoutMillSec float64) map[string]any {

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMillSec)*time.Millisecond)
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
	resp, err := http.DefaultClient.Do(req)
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
	res["header"] = hds
	res["status"] = float64(resp.StatusCode)
	return res
})
