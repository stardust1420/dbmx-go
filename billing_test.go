package dbmx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
)

func TestPlanRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/chat" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"success":false,"message":"plan_required","status_code":402}`))
	}))
	defer srv.Close()

	c := NewClient(Credentials{BaseURL: srv.URL, AccessToken: "token"})
	_, err := c.Chat(context.Background(), ChatReq{})
	if !errors.Is(err, ErrPlanRequired) {
		t.Fatalf("want ErrPlanRequired, got %v", err)
	}
}
