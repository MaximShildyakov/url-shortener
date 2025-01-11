package save

import (
	"errors"
	"net/http"

	"github.com/MaximShildyakov/url-shortener/internal/lib/random"

	"github.com/MaximShildyakov/url-shortener/internal/lib/logger/sl"
	"github.com/MaximShildyakov/url-shortener/internal/storage"
	resp "github.com/MaximShildyakov/url-shortener/internal/lib/api/response"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
)

type Request struct{
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" validate:"omitempty,min=3,max=20"`
}

type Response struct {
	resp.Response
	Alias  string `json:"alias,omitempty"`
}

// TODO: move to config
const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface{
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const fn = "handler.url.save.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil{
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))
			
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().StructPartial(req, "URL"); err != nil {
			log.Error("invalid URL", sl.Err(err))
		
			render.JSON(w, r, resp.Error("field URL is not a valid URL"))
			return
		}

		alias := req.Alias
		if alias == ""{
			alias = random.NewRandomString(aliasLength)
		} else {
			if len(alias) < 3 || len(alias) > 20 {
				log.Error("alias length invalid", slog.String("alias", alias))
		
				render.JSON(w, r, resp.Error("alias length must be between 3 and 20 characters"))
		
				return
			}
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists){
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		if errors.Is(err, storage.ErrAliasExists){
			log.Info("alias already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("alias already exists"))

			return
		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, alias)


	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}