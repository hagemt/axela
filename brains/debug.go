package brains

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func dump(b []byte, re *http.Response) string {
	c, ko := httputil.DumpRequestOut(re.Request, false)
	log.Println("--- sent HTTP:", ko, string(c), string(b))
	d, ko := httputil.DumpResponse(re, true)
	log.Println("--- dump HTTP:", ko, string(d))
	return re.Status
}
