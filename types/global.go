package types

import "net/http"

type LastBlockData struct {
	Height    int
	Timestamp int64
}

type Handler struct {
}

type Server struct {
	httpServer *http.Server
}
