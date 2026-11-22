package mockplatform

import (
	_ "embed"
	"net/http"
)

//go:embed ui.html
var uiPage []byte

// ui serves the customer chat page used for manual local runs.
func (s *Server) ui(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(uiPage)
}
