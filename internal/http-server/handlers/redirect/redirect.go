package redirect

import (
	"crypto/internal/fips140/alias"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.redirect.New"

		log := log.With(
			slog.String("op", op), 
			slog.String("request_id", middleware.GetReqID(r.Context()))
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		resUrl, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, urlGetter.ErrUrlNotFound) {
				log.Info("failed to get URL", slog.String("alias", alias), slog.String("error", err.Error()))
				render.JSON(w, r, response.Error("failed to get URL"))
				return
			}
			log.Info("failed to get URL", slog.String("alias", alias), slog.String("error", err.Error()), "internal error")
				render.JSON(w, r, response.Error("failed to get URL, internal error"))
				return
		}

		log.Info("got url", slog.String("url", resUrl))

		http.Redirect(w, r, resUrl, http.StatusFound)
	}
}
