package expr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Second*3)
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
		timeoutMillSec = 60000
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
	res["header"] = hds
	res["status"] = float64(resp.StatusCode)
	return res
})

func init() {

	type key struct {
		url       string
		ip        string
		sslVerify bool
	}
	httpLib := NewInstanceCache[key, any, *http.Client](func(ctx context.Context, k key, c any) (v *http.Client, err error) {
		cli := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					if k.ip == "" {
						return net.DialTimeout(network, addr, time.Second*3)
					}
					_, port, _ := strings.Cut(addr, ":")
					return net.DialTimeout(network, k.ip+":"+port, time.Second*3)
				},
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !k.sslVerify,
				},
			},
		}
		return cli, nil
	})

	RegisterOptFuncDefine1("curl", func(c *Context, url string, opt *Options) *httpResp {

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		method := opt.GetString("method")
		body := opt.Get("body")
		if method == "" {
			if body != nil {
				method = "POST"
			} else {
				method = "GET"
			}
		}
		timeoutMillSec := opt.GetTimeoutMillDef("timeout", 60000)
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
			timeoutMillSec = 60000 * time.Millisecond
		}
		ctx, cancel := context.WithTimeout(c, timeoutMillSec)
		defer cancel()
		res := &httpResp{}
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bb))

		if err != nil {
			res.Err = err.Error()
			return res
		}
		headers, _ := opt.Get("header").(map[string]any)
		for k, v := range headers {
			req.Header.Set(k, StringOf(v))
		}

		ip := opt.GetString("ip")
		cli, _ := httpLib.Get(c, key{url: url, ip: ip, sslVerify: opt.GetBoolDef("ssl_verify", true)}, nil)
		resp, err := cli.Do(req)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		defer resp.Body.Close()
		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			res.Err = err.Error()
			return res
		}
		res.Body = bs
		hds := map[string]any{}
		for key, val := range resp.Header {
			if len(val) > 0 {
				hds[key] = val[0]
			}
		}
		res.Header = hds
		res.Status = resp.StatusCode
		res.StatusLine = resp.Status
		res.Proto = resp.Proto
		return res
	}, Doc("curl(url) options:{method:'GET' or 'POST'(when body not nil) ,header:{},body:nil, ip:''(force_ip), timeout:60000 (ms)})"))

	SelfDefine1("log", func(ctx *Context, self *httpResp, opt map[string]any) any {
		o := NewOptions(opt)

		if self.Err != nil {
			fmt.Fprintln(os.Stderr, self.Err)
			return nil
		}

		all := o.GetBoolDef("all", false)

		if all || o.GetBoolDef("status", false) {
			fmt.Println(self.Proto, self.StatusLine)
		}
		if all || o.GetBoolDef("header", false) {
			for key, val := range self.Header {
				fmt.Printf("%s: %s\n", key, StringOf(val))
			}
			fmt.Println()

		}
		if all || o.GetBoolDef("body", true) {
			fmt.Printf("%s", ToString(self.Body))
		}
		return nil
	}, WithDoc("opt: {status:0,header:0,body:1}   print the status header and body, only print body by default"))

	SelfDefine0("throw", func(ctx *Context, self *httpResp) *httpResp {
		if self.Err != nil {
			panic(fmt.Sprintf("curl throw err:%v", self.Err))
		}
		if self.Status/100 != 2 {
			panic(fmt.Sprintf("%s\n%v\n%s", self.StatusLine, self.Header, self.Body))
		}
		return self
	}, WithDoc(" panic when failed"))

	SelfDefine0("failed", func(ctx *Context, self *httpResp) any {
		if self.Err != "" {
			return self.Err
		}
		if self.Status/100 != 2 {
			return fmt.Sprintf("%s\n%s", self.StatusLine, self.Body)
		}
		return nil
	}, WithDoc(" return nil when err is nil and status is 200-299 or string of err"))
}

type httpResp struct {
	Proto      string
	StatusLine string
	Err        any
	Body       []byte
	Header     map[string]any
	Status     int
}

func (h *httpResp) GetField(c *Context, key string) any {
	switch key {
	case "body":
		if h.Body == nil {
			return nil
		}
		return h.Body
	case "header":
		if h.Header == nil {
			return nil
		}
		return h.Header
	case "err":
		return h.Err
	case "status":
		return float64(h.Status)
	default:
		return nil
	}
}
