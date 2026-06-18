package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerDeleteUsers_ForbiddenWhenNotDev(t *testing.T) {

	/* TODO: create a real test db in testcontainer or an interface mock.
	 * For now this is fine - even if we would change the platform to "dev" the test would panic/fail since cfg.db is nil.
	 * In real environment such test should never be allowed to exist anyway.
	 */

	cfg := &apiConfig{
		platform: "prod",
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/reset", nil)
	w := httptest.NewRecorder()

	cfg.handlerDeleteUsers(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", res.StatusCode)
	}
}
