package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.redirect.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		resUrl, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrUrlNotFound) {
				log.Info("failed to get URL", slog.String("alias", alias), slog.String("error", err.Error()))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("URL not found"))
				return
			}
			log.Info("failed to get URL",
				slog.String("alias", alias),
				slog.String("error", err.Error()),
				slog.String("type", "internal error"),
			)
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get URL, internal error"))
			return
		}

		log.Info("got url", slog.String("url", resUrl))

		http.Redirect(w, r, resUrl, http.StatusFound)
	}
}
