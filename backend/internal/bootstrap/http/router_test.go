package http

import "testing"

func TestNewRouter_InvalidTrustedProxies_ReturnsError(t *testing.T) {
	_, err := NewRouter([]string{"not-a-cidr-or-ip"})
	if err == nil {
		t.Fatal("expected NewRouter to return error for invalid trusted proxies")
	}
}
