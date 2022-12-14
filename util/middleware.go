package util

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler
type Middlewares []Middleware

func (mws Middlewares) Apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].Apply(mws[0](hdlr))
}
