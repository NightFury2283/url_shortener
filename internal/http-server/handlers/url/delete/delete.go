package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"
)

type UrlDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter UrlDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrUrlNotFound) {
				log.Info("failed to get URL", slog.String("alias", alias), slog.String("error", err.Error()))
				render.JSON(w, r, response.Error("URL not found"))
				return
			}
			log.Info("failed to get URL",
				slog.String("alias", alias),
				slog.String("error", err.Error()),
				slog.String("type", "internal error"),
			)
			render.JSON(w, r, response.Error("failed to get URL, internal error"))
			return
		}
		log.Info("deleted url", slog.String("alias", alias))

		storage.ResponseOK(w, r, alias)
	}
}
