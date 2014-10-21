package autobahn

import (
	"github.com/gopherjs/gopherjs/js"
)

type Transport struct {
	js.Object
}

func (t *Transport) GetType() string {
	return t.Get("info").Get("type").Str()
}

func (t *Transport) GetUrl() string {
	return t.Get("info").Get("url").Str()
}

func (t *Transport) GetProtocol() string {
	return t.Get("info").Get("protocol").Str()
}
