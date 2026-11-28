package dbmx

import (
	"context"
	"io"
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

// TestCheckoutSendsInterval: checkout bodies carry the billing period, and
// GetPrices decodes the server's prices.
func TestCheckoutSendsInterval(t *testing.T) {
	bodies := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies[r.URL.Path] = string(b)
		if r.URL.Path == "/api/v1/billing/prices" {
			_, _ = w.Write([]byte(`{"success":true,"data":{"pro":{"month":{"amount":500,"currency":"USD"},"year":null},"team":{"month":{"amount":750,"currency":"USD"},"year":{"amount":8100,"currency":"USD"}}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"url":"https://pay.test"}}`))
	}))
	defer srv.Close()
	c := NewClient(Credentials{BaseURL: srv.URL, AccessToken: "token"})
	ctx := context.Background()

	if _, err := c.CreateCheckout(ctx, "pro", "month"); err != nil {
		t.Fatal(err)
	}
	if got := bodies["/api/v1/billing/checkout"]; got != `{"plan":"pro","interval":"month"}` {
		t.Errorf("checkout body = %s", got)
	}
	if _, err := c.CreateTeamCheckout(ctx, "o1", "year"); err != nil {
		t.Fatal(err)
	}
	if got := bodies["/api/v1/billing/team/checkout"]; got != `{"org_id":"o1","interval":"year"}` {
		t.Errorf("team checkout body = %s", got)
	}

	p, err := c.GetPrices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p.Pro.Month == nil || p.Pro.Month.Amount != 500 || p.Pro.Year != nil || p.Team.Year == nil || p.Team.Year.Amount != 8100 {
		t.Errorf("prices = %+v", p)
	}
}
