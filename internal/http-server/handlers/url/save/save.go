package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	my_slog "url-shortener/internal/lib/logger/my_slog"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

const (
	aliasLength = 8
)

type UrlSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req storage.Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request body", my_slog.Err(err))
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("request validation failed", my_slog.Err(err))

			render.JSON(w, r, response.Error("validation failed: invalid url format or missing required fields"))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.GenerateRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Warn("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, response.Error("url with this alias already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", my_slog.Err(err))

			render.JSON(w, r, response.Error("internal server error, failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))
		storage.ResponseOK(w, r, alias)
	}
}
