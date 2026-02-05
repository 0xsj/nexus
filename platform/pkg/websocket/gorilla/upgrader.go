package gorilla

import (
	"net/http"

	gorillaWs "github.com/gorilla/websocket"

	ws "github.com/0xsj/nexus/platform/pkg/websocket"
)

// Upgrader wraps the Gorilla WebSocket upgrader with configuration support.
type Upgrader struct {
	upgrader gorillaWs.Upgrader
	cfg      ws.Config
}

// NewUpgrader creates a new Gorilla-based WebSocket upgrader.
func NewUpgrader(cfg ws.Config) *Upgrader {
	u := &Upgrader{
		cfg: cfg,
		upgrader: gorillaWs.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
		},
	}

	if len(cfg.AllowedOrigins) > 0 {
		allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
		for _, origin := range cfg.AllowedOrigins {
			allowed[origin] = struct{}{}
		}
		u.upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			_, ok := allowed[origin]
			return ok
		}
	} else {
		// Allow all origins in development.
		u.upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	return u
}

// Upgrade upgrades an HTTP connection to a WebSocket connection.
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (ws.Connection, error) {
	conn, err := u.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, ws.ErrUpgradeFailed("gorilla.Upgrader.Upgrade", err)
	}
	return NewConnection(conn, u.cfg), nil
}
