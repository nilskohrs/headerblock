package headerblock_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nilskohrs/headerblock"
)

func TestShouldBlockConfiguredHeaders(t *testing.T) {
	cfg := headerblock.CreateConfig()
	cfg.RequestHeaders = append(cfg.RequestHeaders, headerblock.HeaderConfig{
		Name:  "^X-",
		Value: "Evil",
	}, headerblock.HeaderConfig{
		Name: "^Irrelevant",
	}, headerblock.HeaderConfig{
		Value: "ignored$",
	})
	cfg.ResponseHeaders = append(cfg.ResponseHeaders, headerblock.HeaderConfig{
		Name:  "Confusing-For-Client",
		Value: "Weird-Value",
	}, headerblock.HeaderConfig{
		Name: "^Internal-",
	}, headerblock.HeaderConfig{
		Value: "leaking secret$",
	})

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Confusing-For-Client", "Weird-Value")
		rw.Header().Add("Confusing-For-Client", "Normal-Value")
		rw.Header().Add("Internal-Source", "Database")
		rw.Header().Add("Internal-Source", "Cache")
		rw.Header().Add("Internal-State", "Ready")
		rw.Header().Add("Secret-Thingy", "provided: leaking secret")
	})

	handler, err := headerblock.New(ctx, next, cfg, "headerBlock")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/admin/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("X-Permission", "Something Evil in here")
	req.Header.Add("X-Permission", "Something Okay in here")
	req.Header.Add("X-Trace", "All fine")
	req.Header.Add("Irrelevant-Header", "whatever")
	req.Header.Add("Irrelevant-Header", "not relevant")
	req.Header.Add("User-Setting", "things, but ignored")

	handler.ServeHTTP(recorder, req)

	assertHeadersDoNotContain(t, req.Header, "X-Permission", "Something Evil in here")
	assertHeadersDoContain(t, req.Header, "X-Permission", "Something Okay in here")
	assertHeadersDoContain(t, req.Header, "X-Trace", "All fine")
	assertHeadersDoNotContain(t, req.Header, "Irrelevant-Header", "whatever")
	assertHeadersDoNotContain(t, req.Header, "Irrelevant-Header", "not relevant")
	assertHeadersDoNotContain(t, req.Header, "User-Setting", "things, but ignored")

	assertHeadersDoNotContain(t, recorder.Header(), "Confusing-For-Client", "Weird-Value")
	assertHeadersDoContain(t, recorder.Header(), "Confusing-For-Client", "Normal-Value")
	assertHeadersDoNotContain(t, recorder.Header(), "Internal-Source", "Database")
	assertHeadersDoNotContain(t, recorder.Header(), "Internal-Source", "Cache")
	assertHeadersDoNotContain(t, recorder.Header(), "Internal-State", "Ready")
	assertHeadersDoNotContain(t, recorder.Header(), "Secret-Thingy", "provided: leaking secret")
}

func assertHeadersDoNotContain(t *testing.T, headers http.Header, header, value string) {
	t.Helper()
	headerValues := headers.Values(header)
	for _, headerValue := range headerValues {
		if value == headerValue {
			t.Errorf("header `%s` shouldn't contain value `%s` but did", header, value)
		}
	}
}

func assertHeadersDoContain(t *testing.T, headers http.Header, header, value string) {
	t.Helper()
	headerValues := headers.Values(header)
	for _, headerValue := range headerValues {
		if value == headerValue {
			return
		}
	}
	t.Errorf("header `%s` should contain value `%s` but didn't", header, value)
}
