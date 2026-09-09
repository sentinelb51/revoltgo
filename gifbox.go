package revoltgo

//go:generate msgp -tests=false -io=false

import (
	"net/url"
	"strconv"
)

// Gifbox is a service of its own: it holds the key to the GIF provider it
// proxies and takes this session's token in its place.

// GifboxLocaleDefault is sent by a request naming no locale; every route
// requires one.
const GifboxLocaleDefault = "en_US"

// GIF is one result. URL is the page it is published at, which is what a
// message carries; MediaFormats is keyed by the provider's own names.
type GIF struct {
	ID           string              `json:"id"`
	URL          string              `json:"url"`
	MediaFormats map[string]GIFMedia `json:"media_formats"`
}

type GIFMedia struct {
	URL        string `json:"url"`
	Dimensions []int  `json:"dimensions"` // width, then height
}

type GIFPage struct {
	Results []*GIF `json:"results"`
	Next    string `json:"next,omitzero"` // empty on the last page
}

type GIFCategory struct {
	Title string `json:"title"`
	Image string `json:"image"`
}

type GIFSearchParams struct {
	Query  string
	Locale string

	Limit    int    // 0 leaves the service's own default
	Position string // GIFPage.Next

	// IsCategory marks a query picked from GIFCategory rather than typed.
	IsCategory bool
}

func (p GIFSearchParams) values() url.Values {
	values := gifboxValues(p.Locale, p.Limit, p.Position)
	values.Set("query", p.Query)

	if p.IsCategory {
		values.Set("is_category", "true")
	}

	return values
}

type GIFTrendingParams struct {
	Locale string

	Limit    int    // 0 leaves the service's own default
	Position string // GIFPage.Next
}

func (p GIFTrendingParams) values() url.Values {
	return gifboxValues(p.Locale, p.Limit, p.Position)
}

// gifboxValues holds what every Gifbox route takes.
func gifboxValues(locale string, limit int, position string) url.Values {
	values := make(url.Values, 3)

	if locale == "" {
		locale = GifboxLocaleDefault
	}
	values.Set("locale", locale)

	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}

	if position != "" {
		values.Set("position", position)
	}

	return values
}
