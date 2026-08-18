package test

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"nurture/internal/config"
	handlerchat "nurture/internal/handler/chat"
	"nurture/internal/pkg/jwtx"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestChatDirectSessionDeliversMessages(t *testing.T) {
	router, closeServer := newChatSessionServer()
	defer closeServer()

	user1Token := mustChatToken(t, "user-1")
	user2Token := mustChatToken(t, "user-2")

	conn1 := mustDialChat(t, router, "/api/chat/direct", url.Values{
		"token":   {user1Token},
		"user_id": {"user-2"},
	})
	defer conn1.Close()

	conn2 := mustDialChat(t, router, "/api/chat/direct", url.Values{
		"token":   {user2Token},
		"user_id": {"user-1"},
	})
	defer conn2.Close()

	if err := conn1.WriteMessage(websocket.TextMessage, []byte("hello\nthere")); err != nil {
		t.Fatalf("write message: %v", err)
	}

	_, got, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	if string(got) != "hello there" {
		t.Fatalf("got %q, want hello there", got)
	}
}

func TestChatGroupSessionRejectsInvalidMessage(t *testing.T) {
	router, closeServer := newChatSessionServer()
	defer closeServer()

	conn := mustDialChat(t, router, "/api/chat/groups/session", url.Values{
		"token": {mustChatToken(t, "user-1")},
	})
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{")); err != nil {
		t.Fatalf("write message: %v", err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}

	var ack struct {
		Op    string `json:"op"`
		For   string `json:"for"`
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &ack); err != nil {
		t.Fatalf("unmarshal ack: %v", err)
	}
	if ack.Op != "ack" || ack.For != "parse" || ack.Ok || ack.Error != "invalid_json" {
		t.Fatalf("unexpected ack: %+v", ack)
	}
}

func newChatSessionServer() (string, func()) {
	gin.SetMode(gin.TestMode)
	config.Conf.Auth.AccessSecret = "test-secret"
	config.Conf.Auth.AccessExpire = 3600

	handler := handlerchat.NewHandler()
	engine := gin.New()
	engine.GET("/api/chat/direct", handler.OpenDirectSession)
	engine.GET("/api/chat/groups/session", handler.OpenGroupSession)

	server := httptest.NewServer(engine)
	return "ws" + strings.TrimPrefix(server.URL, "http"), server.Close
}

func mustChatToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := jwtx.GenTestToken(userID, jwtx.COMMON_USER)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func mustDialChat(t *testing.T, baseURL string, path string, query url.Values) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(baseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		t.Fatalf("dial %s: %v", path, err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	return conn
}
