package services

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type kitLoggerAdapter struct {
	log zerolog.Logger
}

func (n kitLoggerAdapter) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "(MISSING VALUE)")
	}

	event := n.log.Trace()
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = "(INVALID KEY)"
		}

		if key == zerolog.LevelFieldName {
			continue
		}

		value := keyvals[i+1]
		event.Any(key, value)
	}
	event.Msg("")
	return nil
}

func (s *signalrService) HandleFunc(path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.app.Get(path, adaptor.HTTPHandlerFunc(f))
	s.app.Post(path, adaptor.HTTPHandlerFunc(f))
}

func (s *signalrService) Handle(path string, _ http.Handler) {
	s.app.Get(path, s.AccessTokenMapper, s.auth.Middleware, s.ConnectEndpoint)
}

func (s *signalrService) AccessTokenMapper(ctx *fiber.Ctx) error {
	token := ctx.Query("access_token")
	if token != "" {
		ctx.Request().Header.Set("Authorization", "Bearer "+token)
	}
	return ctx.Next()
}

func (s *signalrService) ConnectEndpoint(ctx *fiber.Ctx) error {
	connectionID := ctx.Query("id")
	if connectionID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing connection id",
		})
	}

	user := GetFromContext(ctx, UserKey)
	if err := s.upgrader().Upgrade(ctx.Context(), s.wsInit(user.ID, connectionID)); err != nil {
		s.log.Error().Err(err).Str("user", user.Name).Msg("Failed to upgrade connection")
		return ctx.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	s.clients.Set(user.ID, connectionID)
	if user.HasRole(models.ViewAllDownloads) {
		//s.Groups().AddToGroup(allDownloadInfoGroup, connectionID)
	}

	return nil
}

func (s *signalrService) upgrader() *websocket.FastHTTPUpgrader {
	var upgrader = websocket.FastHTTPUpgrader{
		EnableCompression: true,
		HandshakeTimeout:  5 * time.Second,
	}

	if config.Development {
		upgrader.CheckOrigin = func(ctx *fasthttp.RequestCtx) bool {
			return true
		}
	}

	return &upgrader
}

func (s *signalrService) wsInit(userId uint, id string) func(conn *websocket.Conn) {
	return func(conn *websocket.Conn) {
		if err := s.server.Serve(newFastHttpConn(conn, id)); err != nil {
			// 1001 Going Away & 1000 Normal Closure
			if strings.Contains(err.Error(), "1001") || strings.Contains(err.Error(), "1000") {
				return
			}

			// Serve returns an error for any close, also intended ones. Let us log in debug, useful info should be
			// on client site. And if signalR is having issues; we'll probably want debug anyway to figure it out.
			s.log.Debug().Err(err).Msg("websocket connection failed or ended")
		} else {
			s.log.Debug().Str("id", id).Msg("websocket connection succeeded")
		}

		s.clients.Delete(userId)
		//s.Groups().RemoveFromGroup(allDownloadInfoGroup, id)
	}
}

func newFastHttpConn(conn *websocket.Conn, id string) *wsFastHttpConn {
	return &wsFastHttpConn{
		conn: conn,
		id:   id,
	}
}

type wsFastHttpConn struct {
	conn  *websocket.Conn
	id    string
	wlock sync.Mutex
	rlock sync.Mutex
}

func (w *wsFastHttpConn) Read(p []byte) (n int, err error) {
	if w.conn == nil {
		return 0, io.EOF
	}

	w.rlock.Lock()
	defer w.rlock.Unlock()
	_, data, err := w.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(p, data), nil
}

func (w *wsFastHttpConn) Write(p []byte) (n int, err error) {
	if w.conn == nil {
		return 0, io.EOF
	}

	w.wlock.Lock()
	defer w.wlock.Unlock()
	if err = w.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *wsFastHttpConn) Context() context.Context {
	return context.Background()
}

func (w *wsFastHttpConn) ConnectionID() string {
	return w.id
}

func (w *wsFastHttpConn) SetConnectionID(id string) {
	w.id = id
}
