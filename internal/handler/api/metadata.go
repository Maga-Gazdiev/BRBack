package api

import (
	"encoding/xml"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

// nginx includes these tags while serving index.html, so social crawlers see
// the current CMS metadata without JavaScript or rebuilding the frontend.
var metadataTemplate = template.Must(template.New("metadata").Parse(`<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<meta property="og:type" content="website">
<meta property="og:locale" content="ru_RU">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
{{if .URL}}<link rel="canonical" href="{{.URL}}"><meta property="og:url" content="{{.URL}}">{{end}}
{{if .Image}}<meta property="og:image" content="{{.Image}}">{{end}}
{{if .Admin}}<meta name="robots" content="noindex,nofollow">{{end}}`))

func (h *Handler) metadata(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.Get(r.Context())
	if err != nil {
		h.fail(w, err)
		return
	}
	imageURL := ""
	if c.Settings.SiteURL != "" {
		base, _ := url.Parse(c.Settings.SiteURL)
		ref, _ := url.Parse(c.Hero.Image)
		if base != nil && ref != nil {
			imageURL = base.ResolveReference(ref).String()
		}
	}
	admin := strings.HasPrefix(r.Header.Get("X-Page-Path"), "/admin")
	title := c.Settings.Title
	if admin {
		title = "Редактор — " + c.Brand
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_ = metadataTemplate.Execute(w, struct {
		Title, Description, URL, Image string
		Admin                          bool
	}{title, c.Settings.Description, c.Settings.SiteURL, imageURL, admin})
}
func (h *Handler) robots(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.Get(r.Context())
	if err != nil {
		h.fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /admin\nDisallow: /api/\n"))
	if c.Settings.SiteURL != "" {
		_, _ = w.Write([]byte("Sitemap: " + strings.TrimRight(c.Settings.SiteURL, "/") + "/sitemap.xml\n"))
	}
}
func (h *Handler) sitemap(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.Get(r.Context())
	if err != nil {
		h.fail(w, err)
		return
	}
	if c.Settings.SiteURL == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>`))
	_ = xml.EscapeText(w, []byte(c.Settings.SiteURL))
	_, _ = w.Write([]byte(`</loc></url></urlset>`))
}
