package main

import (
	"dobrikov91/shubert/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestWebServer(t *testing.T) *WebServer {
	t.Helper()

	config, err := model.NewConfig(t.TempDir() + "/config.json")
	if err != nil {
		t.Fatalf("can't create config: %v", err)
	}

	c := &Controller{Config: config, Port: "0"}
	web := &WebServer{c: c, broadcast: make(chan model.Commands)}

	// drain broadcasts like the real server does via Run(), so handlers
	// that send to web.broadcast don't block forever in tests
	go func() {
		for range web.broadcast {
		}
	}()

	return web
}

func postForm(t *testing.T, handler http.HandlerFunc, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestCommandFromForm(t *testing.T) {
	form := url.Values{
		"Device":  {"midi1"},
		"Channel": {"1"},
		"Key":     {"2"},
		"Alias":   {"my alias"},
		"Trigger": {"OnPress"},
		"Command": {"echo hi"},
		"Timeout": {"500"},
	}

	cmd := commandFromForm(form, 0)

	if cmd.Event.Device != "midi1" || cmd.Event.Channel != 1 || cmd.Event.Key != 2 {
		t.Errorf("unexpected event: %+v", cmd.Event)
	}
	if cmd.Alias != "my alias" || cmd.Trigger != "OnPress" || cmd.Command != "echo hi" {
		t.Errorf("unexpected command fields: %+v", cmd)
	}
	if cmd.TimeoutMs != 500 {
		t.Errorf("expected TimeoutMs 500, got %d", cmd.TimeoutMs)
	}
}

func TestAtoi(t *testing.T) {
	if got := atoi("42"); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
	if got := atoi("not a number"); got != 0 {
		t.Errorf("expected 0 for invalid input, got %d", got)
	}
}

func TestHandleSave(t *testing.T) {
	web := newTestWebServer(t)

	form := url.Values{
		"Device":  {"midi1", "midi1"},
		"Channel": {"1", "2"},
		"Key":     {"10", "20"},
		"Alias":   {"first", "second"},
		"Trigger": {"OnPress", "OnRelease"},
		"Command": {"echo one", "echo two"},
		"Timeout": {"0", "0"},
	}

	rec := postForm(t, web.handleSave, form)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}

	if len(web.c.Config.Data.Commands) != 2 {
		t.Fatalf("expected 2 commands in memory, got %d", len(web.c.Config.Data.Commands))
	}
	if web.c.Config.Data.Commands[1].Alias != "second" {
		t.Errorf("expected second command alias 'second', got %q", web.c.Config.Data.Commands[1].Alias)
	}

	// verify it was actually persisted to disk, not just in memory
	reloaded, err := model.NewConfig(web.c.Config.FilePath)
	if err != nil {
		t.Fatalf("can't reload saved config: %v", err)
	}
	if len(reloaded.Data.Commands) != 2 {
		t.Errorf("expected 2 commands on disk, got %d", len(reloaded.Data.Commands))
	}
}

func TestHandleSaveReplacesExistingCommands(t *testing.T) {
	web := newTestWebServer(t)
	web.c.Config.AddCommand(model.Command{Event: model.Event{Device: "stale", Channel: 1, Key: 1}})

	// saving an empty form should clear the previous command, not append
	rec := postForm(t, web.handleSave, url.Values{})

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if len(web.c.Config.Data.Commands) != 0 {
		t.Errorf("expected save with no rows to clear commands, got %d", len(web.c.Config.Data.Commands))
	}
}

func TestHandleDeleteValidIndex(t *testing.T) {
	web := newTestWebServer(t)
	web.c.Config.AddCommand(model.Command{Alias: "keep-me-not"})
	web.c.Config.AddCommand(model.Command{Alias: "keep-me"})

	rec := postForm(t, web.handleDelete, url.Values{"index": {"0"}})

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if len(web.c.Config.Data.Commands) != 1 {
		t.Fatalf("expected 1 command left, got %d", len(web.c.Config.Data.Commands))
	}
	if web.c.Config.Data.Commands[0].Alias != "keep-me" {
		t.Errorf("deleted the wrong command: %+v", web.c.Config.Data.Commands[0])
	}
}

func TestHandleDeleteInvalidIndex(t *testing.T) {
	web := newTestWebServer(t)
	web.c.Config.AddCommand(model.Command{Alias: "only-one"})

	for _, idx := range []string{"-1", "1"} {
		rec := postForm(t, web.handleDelete, url.Values{"index": {idx}})

		if rec.Code != http.StatusBadRequest {
			t.Errorf("index %q: expected %d, got %d", idx, http.StatusBadRequest, rec.Code)
		}
		if len(web.c.Config.Data.Commands) != 1 {
			t.Errorf("index %q: command list should be unchanged, got %d commands", idx, len(web.c.Config.Data.Commands))
		}
	}
}

func TestHandleDeleteNonNumericIndexDefaultsToZero(t *testing.T) {
	// atoi() silently treats a non-numeric index as 0 rather than
	// rejecting it, so this deletes the first command instead of
	// erroring — documenting the current (surprising) behavior.
	web := newTestWebServer(t)
	web.c.Config.AddCommand(model.Command{Alias: "only-one"})

	rec := postForm(t, web.handleDelete, url.Values{"index": {"notanumber"}})

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if len(web.c.Config.Data.Commands) != 0 {
		t.Errorf("expected the only command to be deleted, got %d left", len(web.c.Config.Data.Commands))
	}
}
