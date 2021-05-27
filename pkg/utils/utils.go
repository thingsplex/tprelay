package utils

import (
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func GenerateRandomNumber() int64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Int63()
}

func TunFrameToHttpReq(tunFrame *tunframe.TunnelFrame) *http.Request {
	headers := http.Header{}
	if tunFrame.Headers != nil {
		for k,v := range tunFrame.Headers {
			headers[k] = v.Items
		}
	}
	r := http.Request{Method: tunFrame.ReqMethod,Header: headers}
	url,err := url.Parse(tunFrame.ReqUrl)
	if err == nil {
		r.URL = url
	}
	r.RequestURI = tunFrame.ReqUrl
	return &r
}
