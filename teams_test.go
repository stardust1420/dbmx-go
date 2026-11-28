package dbmx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
)

func TestAPIErrorMessage(t *testing.T) {
	const msg = "Trial is limited to 5 members — subscribe to add more."
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"success":false,"message":"` + msg + `","status_code":409}`))
	}))
	defer srv.Close()

	c := NewClient(Credentials{BaseURL: srv.URL, AccessToken: "token"})
	_, err := c.InviteMember(context.Background(), "org", "a@b.co")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %v", err)
	}
	if apiErr.Status != http.StatusConflict || apiErr.Message != msg {
		t.Fatalf("got %d %q", apiErr.Status, apiErr.Message)
	}
}
