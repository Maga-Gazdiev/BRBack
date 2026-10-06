// Package content owns validation and editing rules. Storage is injected.
package content

import (
	"context"
	"errors"
	"fmt"
	"m96/internal/model"
	"net/url"
	"regexp"
	"strings"
)

var ErrConflict = errors.New("content was changed by another editor")
var ErrInvalid = errors.New("invalid content")

type Repository interface {
	Get(context.Context) (model.Content, error)
	Save(context.Context, model.Content) (model.Content, error)
}
type Service struct{ repository Repository }

func New(r Repository) *Service                                   { return &Service{repository: r} }
func (s *Service) Get(ctx context.Context) (model.Content, error) { return s.repository.Get(ctx) }
func (s *Service) Update(ctx context.Context, c model.Content) (model.Content, error) {
	if err := Validate(c); err != nil {
		return model.Content{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	return s.repository.Save(ctx, c)
}
func webURL(s string) bool {
	u, e := url.Parse(s)
	return e == nil && u.Scheme == "https" && u.Host != "" && u.User == nil
}
func imageURL(s string) bool {
	return webURL(s) || (strings.HasPrefix(s, "/images/") || strings.HasPrefix(s, "/uploads/")) && !strings.Contains(s, "..") && !strings.ContainsAny(s, "?#\\")
}
func Validate(c model.Content) error {
	if c.Revision < 1 || strings.TrimSpace(c.Brand) == "" || strings.TrimSpace(c.Hero.Title) == "" || strings.TrimSpace(c.Settings.Title) == "" {
		return errors.New("заполните название, заголовок и SEO title")
	}
	if len(c.Services) == 0 || len(c.Services) > 100 || len(c.Gallery) > 30 || len(c.Masters) > 30 || len(c.Benefits) > 20 {
		return errors.New("допустимо 1–100 услуг и до 30 фотографий/мастеров, до 20 преимуществ")
	}
	seen := map[string]bool{}
	for _, s := range c.Services {
		if s.ID == "" || seen[s.ID] || strings.TrimSpace(s.Name) == "" || strings.TrimSpace(s.Category) == "" || s.Price < 0 || s.Price > 1000000 || s.Duration < 1 || s.Duration > 1440 {
			return errors.New("проверьте ID, название, категорию, цену и длительность услуг")
		}
		seen[s.ID] = true
	}
	images := []string{c.Hero.Image, c.About.Image}
	for _, p := range c.Gallery {
		images = append(images, p.Image)
	}
	for _, m := range c.Masters {
		if strings.TrimSpace(m.Name) == "" {
			return errors.New("укажите имя мастера")
		}
		images = append(images, m.Image)
	}
	for _, s := range images {
		if !imageURL(s) {
			return errors.New("фотография: требуется HTTPS URL или путь /images/, /uploads/")
		}
	}
	for _, s := range []string{c.Contact.Telegram, c.Contact.VK, c.Contact.RouteURL, c.Settings.SiteURL} {
		if s != "" && !webURL(s) {
			return errors.New("ссылки должны начинаться с https://")
		}
	}
	if s := c.Contact.MapURL; s != "" {
		u, e := url.Parse(s)
		if e != nil || u.Scheme != "https" || u.Host != "yandex.ru" || !strings.HasPrefix(u.Path, "/map-widget/") {
			return errors.New("карта: используйте https://yandex.ru/map-widget/...")
		}
	}
	if s := c.Settings.BookingURL; s != "" {
		u, e := url.Parse(s)
		if e != nil || !webURL(s) || !bookingHost(u.Hostname()) {
			return errors.New("запись: используйте HTTPS ссылку YCLIENTS")
		}
	}
	if s := c.Settings.MetrikaID; s != "" && !regexp.MustCompile(`^[0-9]{1,12}$`).MatchString(s) {
		return errors.New("ID Метрики должен содержать только цифры")
	}
	return nil
}

func bookingHost(host string) bool {
	host = strings.ToLower(host)
	for _, domain := range []string{"yclients.com", "yclients.ru"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}
