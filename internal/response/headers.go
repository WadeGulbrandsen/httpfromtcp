package response

import (
	"fmt"

	"github.com/WadeGulbrandsen/httpfromtcp/internal/headers"
)

func GetDefualtHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}
