package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clientIP(t *testing.T, trustedProxies []string, remoteAddr string, headers map[string]string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r, err := newEngine(trustedProxies)
	require.NoError(t, err)

	r.GET("/ip", func(c *gin.Context) { c.String(nethttp.StatusOK, c.ClientIP()) })

	req := httptest.NewRequest(nethttp.MethodGet, "/ip", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec.Body.String()
}

func TestNewEngine_ClientIP(t *testing.T) {
	subnet := []string{"172.30.0.0/24"}

	tests := []struct {
		name           string
		trustedProxies []string
		remoteAddr     string
		headers        map[string]string
		want           string
	}{
		{
			name:           "trusted proxy: forwarded address is used",
			trustedProxies: subnet,
			remoteAddr:     "172.30.0.5:41000",
			headers:        map[string]string{"X-Real-IP": "203.0.113.7"},
			want:           "203.0.113.7",
		},
		{
			name:           "trusted proxy: X-Forwarded-For is honoured too",
			trustedProxies: subnet,
			remoteAddr:     "172.30.0.5:41000",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.7"},
			want:           "203.0.113.7",
		},
		{
			name:           "untrusted peer cannot forge its address",
			trustedProxies: subnet,
			remoteAddr:     "198.51.100.9:41000",
			headers:        map[string]string{"X-Real-IP": "203.0.113.7", "X-Forwarded-For": "203.0.113.7"},
			want:           "198.51.100.9",
		},
		{
			name:           "trusted proxy without headers: connection address",
			trustedProxies: subnet,
			remoteAddr:     "172.30.0.5:41000",
			want:           "172.30.0.5",
		},
		{
			name:           "empty list: headers are never believed",
			trustedProxies: nil,
			remoteAddr:     "172.30.0.5:41000",
			headers:        map[string]string{"X-Real-IP": "203.0.113.7"},
			want:           "172.30.0.5",
		},
		{
			name:           "blank entries are ignored",
			trustedProxies: []string{" ", ""},
			remoteAddr:     "172.30.0.5:41000",
			headers:        map[string]string{"X-Real-IP": "203.0.113.7"},
			want:           "172.30.0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, clientIP(t, tt.trustedProxies, tt.remoteAddr, tt.headers))
		})
	}
}

func TestNewEngine_RejectsInvalidProxy(t *testing.T) {
	engine, err := newEngine([]string{"not-a-cidr"})

	assert.Error(t, err)
	assert.Nil(t, engine)
}
