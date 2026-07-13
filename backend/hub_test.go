package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// testDB and testRouter are shared across every test in the package - built
// once in TestMain against the real database described by .env, exactly the
// way main() builds them for the live server.
var (
	testDB     *sql.DB
	testRouter *mux.Router
)

// TestMain starts the hub's dispatch goroutine once for the whole test
// binary. Without this, HandleWebSocket's `hub.register <- client` and any
// SendNotificationTo* call block forever - there'd be nothing on the other
// end of the channel to receive them, since main() (which normally starts
// the hub) never runs under `go test`. It also opens the real database
// connection and builds the real router so integration tests exercise the
// exact same wiring the live server uses.
func TestMain(m *testing.M) {
	startHub()

	_ = godotenv.Load() // best effort, same as main()

	db, err := ConnectDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "NOTE: skipping DB-backed tests, could not connect: %v\n", err)
	} else {
		testDB = db
		testRouter = SetupRouter(db)
	}

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

// wsTestServer spins up a real HTTP server exposing exactly the WS upgrade
// route, so these tests exercise the actual HandleWebSocket + hub goroutine
// pipeline end to end rather than mocking any of it.
func wsTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", HandleWebSocket)
	return httptest.NewServer(mux)
}

func dialWS(t *testing.T, serverURL, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws?token=" + token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	return conn
}

func readWithTimeout(conn *websocket.Conn, timeout time.Duration) (wsMessage, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	var msg wsMessage
	err := conn.ReadJSON(&msg)
	return msg, err
}

// TC-WS-01: the upgrade must be refused for an unauthenticated request -
// HandleWebSocket's own auth check, not just JWTMiddleware (which this
// route deliberately bypasses).
func TestHandleWebSocket_RejectsMissingToken(t *testing.T) {
	server := wsTestServer(t)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected the handshake to fail without a token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		status := "no response"
		if resp != nil {
			status = resp.Status
		}
		t.Fatalf("expected 401 Unauthorized, got %s", status)
	}
}

// TC-WS-02: the upgrade must also be refused for a garbage/tampered token.
func TestHandleWebSocket_RejectsInvalidToken(t *testing.T) {
	server := wsTestServer(t)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?token=not-a-real-jwt"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected the handshake to fail with an invalid token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		status := "no response"
		if resp != nil {
			status = resp.Status
		}
		t.Fatalf("expected 401 Unauthorized, got %s", status)
	}
}

// TC-WS-03: a notification sent to a specific userID reaches only that
// user's connection - the hub's core per-user routing guarantee, which is
// what keeps one employee from seeing another employee's notifications.
func TestHub_SendNotificationToUser_OnlyReachesTargetUser(t *testing.T) {
	server := wsTestServer(t)
	defer server.Close()

	tokenA, _ := GenerateJWT(9001, "userA@example.com", "employee")
	tokenB, _ := GenerateJWT(9002, "userB@example.com", "employee")

	connA := dialWS(t, server.URL, tokenA)
	defer connA.Close()
	connB := dialWS(t, server.URL, tokenB)
	defer connB.Close()

	time.Sleep(150 * time.Millisecond) // let both registrations reach the hub goroutine

	SendNotificationToUser(9001, Notification{ID: 1, Title: "For A only", Message: "hello"})

	msgA, err := readWithTimeout(connA, 2*time.Second)
	if err != nil {
		t.Fatalf("user A did not receive the notification targeted at them: %v", err)
	}
	if msgA.Type != "notification" {
		t.Errorf("message type = %q, want notification", msgA.Type)
	}

	// user B must NOT receive a notification that was targeted at user A
	if _, err := readWithTimeout(connB, 300*time.Millisecond); err == nil {
		t.Fatal("user B unexpectedly received a notification targeted at user A")
	}
}

// TC-WS-04: a notification sent to a role reaches every connected client
// with that role, and nobody with a different role - this is what backs
// "HR broadcasts to all employees" and "all HR managers hear about a new
// leave request" without leaking either one to the wrong audience.
func TestHub_SendNotificationToRole_OnlyReachesThatRole(t *testing.T) {
	server := wsTestServer(t)
	defer server.Close()

	hrToken, _ := GenerateJWT(9101, "hr@example.com", "hr_manager")
	empToken, _ := GenerateJWT(9102, "emp@example.com", "employee")

	hrConn := dialWS(t, server.URL, hrToken)
	defer hrConn.Close()
	empConn := dialWS(t, server.URL, empToken)
	defer empConn.Close()

	time.Sleep(150 * time.Millisecond)

	SendNotificationToRole("hr_manager", Notification{ID: 2, Title: "HR only", Message: "announcement"})

	msg, err := readWithTimeout(hrConn, 2*time.Second)
	if err != nil {
		t.Fatalf("HR connection did not receive the role-targeted notification: %v", err)
	}
	if msg.Type != "notification" {
		t.Errorf("message type = %q, want notification", msg.Type)
	}

	if _, err := readWithTimeout(empConn, 300*time.Millisecond); err == nil {
		t.Fatal("employee connection unexpectedly received an HR-only notification")
	}
}

// TC-WS-05: one client disconnecting must not affect delivery to others and
// must not crash the hub goroutine - if it did, every subsequent WS test
// (and every real user's live connection) would silently stop working.
func TestHub_SurvivesClientDisconnect(t *testing.T) {
	server := wsTestServer(t)
	defer server.Close()

	tokenA, _ := GenerateJWT(9201, "gone@example.com", "employee")
	tokenB, _ := GenerateJWT(9202, "stays@example.com", "employee")

	connA := dialWS(t, server.URL, tokenA)
	connB := dialWS(t, server.URL, tokenB)
	defer connB.Close()

	time.Sleep(150 * time.Millisecond)
	connA.Close() // abrupt disconnect
	time.Sleep(150 * time.Millisecond) // let the hub goroutine process the unregister

	SendNotificationToUser(9202, Notification{ID: 3, Title: "Still works", Message: "after a peer disconnected"})

	msg, err := readWithTimeout(connB, 2*time.Second)
	if err != nil {
		t.Fatalf("surviving connection did not receive its notification after a peer disconnected: %v", err)
	}
	if msg.Payload == nil {
		t.Error("expected a non-nil notification payload")
	}
}

// TC-WS-06: many clients connecting and being messaged concurrently must
// not race or deadlock - this is the property the whole register/unregister/
// direct channel design (instead of a locked map) exists to guarantee.
func TestHub_ConcurrentClientsNoRace(t *testing.T) {
	server := wsTestServer(t)
	defer server.Close()

	const numClients = 20
	conns := make([]*websocket.Conn, numClients)
	baseUserID := 9300

	for i := 0; i < numClients; i++ {
		token, _ := GenerateJWT(baseUserID+i, "concurrent@example.com", "employee")
		conns[i] = dialWS(t, server.URL, token)
	}
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()

	time.Sleep(200 * time.Millisecond)

	done := make(chan struct{})
	for i := 0; i < numClients; i++ {
		go func(userID int) {
			SendNotificationToUser(userID, Notification{ID: userID, Title: "concurrent", Message: "test"})
			done <- struct{}{}
		}(baseUserID + i)
	}
	for i := 0; i < numClients; i++ {
		<-done
	}

	for i, c := range conns {
		if _, err := readWithTimeout(c, 2*time.Second); err != nil {
			t.Errorf("client %d did not receive its concurrently-sent notification: %v", i, err)
		}
	}
}
