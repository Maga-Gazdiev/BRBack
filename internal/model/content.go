// Package model defines content independently of transport and storage.
package model

type Content struct {
	Revision int       `json:"revision"`
	Brand    string    `json:"brand"`
	City     string    `json:"city"`
	Hero     Hero      `json:"hero"`
	About    About     `json:"about"`
	Services []Service `json:"services"`
	Gallery  []Photo   `json:"gallery"`
	Benefits []Benefit `json:"benefits"`
	Masters  []Master  `json:"masters"`
	Contact  Contact   `json:"contact"`
	Settings Settings  `json:"settings"`
}
type Hero struct {
	Eyebrow     string `json:"eyebrow"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}
type About struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}
type Service struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Duration    int    `json:"duration"`
}
type Photo struct {
	Image   string `json:"image"`
	Caption string `json:"caption"`
}
type Benefit struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
type Master struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description"`
	Image       string `json:"image"`
}
type Contact struct {
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Hours    string `json:"hours"`
	Telegram string `json:"telegram"`
	VK       string `json:"vk"`
	MapURL   string `json:"mapUrl"`
	RouteURL string `json:"routeUrl"`
}
type Settings struct {
	BookingURL  string `json:"bookingUrl"`
	MetrikaID   string `json:"metrikaId"`
	SiteURL     string `json:"siteUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PrivacyText string `json:"privacyText"`
}
