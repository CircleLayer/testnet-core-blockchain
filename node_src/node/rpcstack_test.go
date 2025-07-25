// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
// Enhanced blockchain implementation by Circle Layer <https://circlelayer.com>

package node

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/internal/testlog"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestCorsHandler makes sure CORS are properly handled on the http server.
func TestCorsHandler(t *testing.T) {
	srv := createAndStartServer(t, &httpConfig{CorsAllowedOrigins: []string{"test", "test.com"}}, false, &wsConfig{})
	defer srv.stop()
	url := "http://" + srv.listenAddr()

	resp := rpcRequest(t, url, "origin", "test.com")
	assert.Equal(t, "test.com", resp.Header.Get("Access-Control-Allow-Origin"))

	resp2 := rpcRequest(t, url, "origin", "bad")
	assert.Equal(t, "", resp2.Header.Get("Access-Control-Allow-Origin"))
}

// TestVhosts makes sure vhosts are properly handled on the http server.
func TestVhosts(t *testing.T) {
	srv := createAndStartServer(t, &httpConfig{Vhosts: []string{"test"}}, false, &wsConfig{})
	defer srv.stop()
	url := "http://" + srv.listenAddr()

	resp := rpcRequest(t, url, "host", "test")
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp2 := rpcRequest(t, url, "host", "bad")
	assert.Equal(t, resp2.StatusCode, http.StatusForbidden)
}

type originTest struct {
	spec    string
	expOk   []string
	expFail []string
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
// Copied over from flags.go
func splitAndTrim(input string) (ret []string) {
	l := strings.Split(input, ",")
	for _, r := range l {
		r = strings.TrimSpace(r)
		if len(r) > 0 {
			ret = append(ret, r)
		}
	}
	return ret
}

// TestWebsocketOrigins makes sure the websocket origins are properly handled on the websocket server.
func TestWebsocketOrigins(t *testing.T) {
	tests := []originTest{
		{
			spec: "*", // allow all
			expOk: []string{"", "http://test", "https://test", "http://test:8540", "https://test:8540",
				"http://test.com", "https://foo.test", "http://testa", "http://atestb:8540", "https://atestb:8540"},
		},
		{
			spec:    "test",
			expOk:   []string{"http://test", "https://test", "http://test:8540", "https://test:8540"},
			expFail: []string{"http://test.com", "https://foo.test", "http://testa", "http://atestb:8540", "https://atestb:8540"},
		},
		// scheme tests
		{
			spec:  "https://test",
			expOk: []string{"https://test", "https://test:9999"},
			expFail: []string{
				"test",                                // no scheme, required by spec
				"http://test",                         // wrong scheme
				"http://test.foo", "https://a.test.x", // subdomain variatoins
				"http://testx:8540", "https://xtest:8540"},
		},
		// ip tests
		{
			spec:  "https://12.34.56.78",
			expOk: []string{"https://12.34.56.78", "https://12.34.56.78:8540"},
			expFail: []string{
				"http://12.34.56.78",     // wrong scheme
				"http://12.34.56.78:443", // wrong scheme
				"http://1.12.34.56.78",   // wrong 'domain name'
				"http://12.34.56.78.a",   // wrong 'domain name'
				"https://87.65.43.21", "http://87.65.43.21:8540", "https://87.65.43.21:8540"},
		},
		// port tests
		{
			spec:  "test:8540",
			expOk: []string{"http://test:8540", "https://test:8540"},
			expFail: []string{
				"http://test", "https://test", // spec says port required
				"http://test:8541", "https://test:8541", // wrong port
				"http://bad", "https://bad", "http://bad:8540", "https://bad:8540"},
		},
		// scheme and port
		{
			spec:  "https://test:8540",
			expOk: []string{"https://test:8540"},
			expFail: []string{
				"https://test",                          // missing port
				"http://test",                           // missing port, + wrong scheme
				"http://test:8540",                      // wrong scheme
				"http://test:8541", "https://test:8541", // wrong port
				"http://bad", "https://bad", "http://bad:8540", "https://bad:8540"},
		},
		// several allowed origins
		{
			spec: "localhost,http://127.0.0.1",
			expOk: []string{"localhost", "http://localhost", "https://localhost:8443",
				"http://127.0.0.1", "http://127.0.0.1:8080"},
			expFail: []string{
				"https://127.0.0.1", // wrong scheme
				"http://bad", "https://bad", "http://bad:8540", "https://bad:8540"},
		},
	}
	for _, tc := range tests {
		srv := createAndStartServer(t, &httpConfig{}, true, &wsConfig{Origins: splitAndTrim(tc.spec)})
		url := fmt.Sprintf("ws://%v", srv.listenAddr())
		for _, origin := range tc.expOk {
			if err := wsRequest(t, url, origin); err != nil {
				t.Errorf("spec '%v', origin '%v': expected ok, got %v", tc.spec, origin, err)
			}
		}
		for _, origin := range tc.expFail {
			if err := wsRequest(t, url, origin); err == nil {
				t.Errorf("spec '%v', origin '%v': expected not to allow,  got ok", tc.spec, origin)
			}
		}
		srv.stop()
	}
}

// TestIsWebsocket tests if an incoming websocket upgrade request is handled properly.
func TestIsWebsocket(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)

	assert.False(t, isWebsocket(r))
	r.Header.Set("upgrade", "websocket")
	assert.False(t, isWebsocket(r))
	r.Header.Set("connection", "upgrade")
	assert.True(t, isWebsocket(r))
	r.Header.Set("connection", "upgrade,keep-alive")
	assert.True(t, isWebsocket(r))
	r.Header.Set("connection", " UPGRADE,keep-alive")
	assert.True(t, isWebsocket(r))
}

func Test_checkPath(t *testing.T) {
	tests := []struct {
		req      *http.Request
		prefix   string
		expected bool
	}{
		{
			req:      &http.Request{URL: &url.URL{Path: "/test"}},
			prefix:   "/test",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/testing"}},
			prefix:   "/test",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/"}},
			prefix:   "/test",
			expected: false,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/fail"}},
			prefix:   "/test",
			expected: false,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/"}},
			prefix:   "",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/fail"}},
			prefix:   "",
			expected: false,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/"}},
			prefix:   "/",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/testing"}},
			prefix:   "/",
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, checkPath(tt.req, tt.prefix))
		})
	}
}

func createAndStartServer(t *testing.T, conf *httpConfig, ws bool, wsConf *wsConfig) *httpServer {
	t.Helper()

	srv := newHTTPServer(testlog.Logger(t, log.LvlDebug), rpc.DefaultHTTPTimeouts)
	assert.NoError(t, srv.enableRPC(nil, *conf))
	if ws {
		assert.NoError(t, srv.enableWS(nil, *wsConf))
	}
	assert.NoError(t, srv.setListenAddr("localhost", 0))
	assert.NoError(t, srv.start())
	return srv
}

// wsRequest attempts to open a WebSocket connection to the given URL.
func wsRequest(t *testing.T, url, browserOrigin string) error {
	t.Helper()
	t.Logf("checking WebSocket on %s (origin %q)", url, browserOrigin)

	headers := make(http.Header)
	if browserOrigin != "" {
		headers.Set("Origin", browserOrigin)
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if conn != nil {
		conn.Close()
	}
	return err
}

// rpcRequest performs a JSON-RPC request to the given URL.
func rpcRequest(t *testing.T, url string, extraHeaders ...string) *http.Response {
	t.Helper()

	// Create the request.
	body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"rpc_modules","params":[]}`))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatal("could not create http request:", err)
	}
	req.Header.Set("content-type", "application/json")

	// Apply extra headers.
	if len(extraHeaders)%2 != 0 {
		panic("odd extraHeaders length")
	}
	for i := 0; i < len(extraHeaders); i += 2 {
		key, value := extraHeaders[i], extraHeaders[i+1]
		if strings.ToLower(key) == "host" {
			req.Host = value
		} else {
			req.Header.Set(key, value)
		}
	}

	// Perform the request.
	t.Logf("checking RPC/HTTP on %s %v", url, extraHeaders)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
