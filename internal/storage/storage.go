package storage

import (
	"errors"
	"net/http"
	"url-shortener/internal/lib/api/response"

	"github.com/go-chi/render"
)

var (
	ErrUrlNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
)

type Request struct {
	URL   string `json:"url" validate:"required,url"` //validate for validator lib: go-playground/validator/v10
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

func ResponseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Alias:    alias,
	})
}
