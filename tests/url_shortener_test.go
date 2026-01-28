package tests

import (
	"net/url"
	"testing"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	gofakeit "github.com/brianvoe/gofakeit/v6"
	he "github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := he.Default(t, u.String())

	e.POST("/url").
		WithJSON(storage.Request{
			URL:   gofakeit.URL(),
			Alias: random.GenerateRandomString(10),
		}).
		WithBasicAuth("admin", "password123").
		Expect().
		Status(200).
		JSON().
		Object().
		ContainsKey("alias")
}
