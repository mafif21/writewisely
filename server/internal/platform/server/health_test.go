package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"writewisely/server/internal/common/constants"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   HealthResponse
	}{
		{
			name:       "db up",
			wantStatus: fiber.StatusOK,
			wantBody:   HealthResponse{Status: constants.HealthStatusOK, DB: constants.DBStatusUp},
		},
		{
			name:       "db down",
			pingErr:    errors.New("connection refused"),
			wantStatus: fiber.StatusServiceUnavailable,
			wantBody:   HealthResponse{Status: constants.HealthStatusDegraded, DB: constants.DBStatusDown},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(fakePinger{err: tt.pingErr})

			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/health", nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			var got HealthResponse
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got != tt.wantBody {
				t.Errorf("body = %+v, want %+v", got, tt.wantBody)
			}
		})
	}
}
