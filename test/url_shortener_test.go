package tests

import (
	// "path"
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"github.com/MaximShildyakov/url-shortener/internal/http-server/handlers/url/save"
	"github.com/MaximShildyakov/url-shortener/internal/lib/api"
	"github.com/MaximShildyakov/url-shortener/internal/lib/random"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth("myuser", "mypass").
		Expect().
		Status(200).
		JSON().
		Object().
		ContainsKey("alias")


}

//nolint:funlen
func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
		{
			name:  "Duplicate Alias",
			url:   gofakeit.URL(),
			alias: "duplicateAlias",
			error: "alias already exists in the database",
		},
		{
			name:  "Empty URL",
			url:   "",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Long Alias",
			url:   gofakeit.URL(),
			alias: gofakeit.LetterN(256),
			error: "alias length must be between 3 and 20 characters",
		},
		{
			name:  "Long URL",
			url:   gofakeit.LetterN(2049),
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// Save

			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("myuser", "mypass").
				Expect().
				Status(http.StatusOK).
				JSON().
				Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")

				resp.Value("error").String().IsEqual(tc.error)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()

				alias = resp.Value("alias").String().Raw()
			}

			// Redirect

			testRedirect(t, alias, tc.url)

			// Remove

			// reqDel := e.DELETE("/"+path.Join("url", alias)).
			// 	WithBasicAuth("myuser", "uXhass").
			// 	Expect().Status(http.StatusOK).
			// 	JSON().Object()
			// reqDel.Value("status").String().IsEqual("OK")

			// // Redirect again 

			// testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) { 
	u := url. URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,

	}

	_, err := api.GetRedirect(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}