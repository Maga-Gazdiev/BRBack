package content

import "testing"

func TestBookingHost(t *testing.T) {
	for host, want := range map[string]bool{
		"yclients.ru": true, "n123.yclients.ru": true, "n123.yclients.com": true,
		"N123.YCLIENTS.RU": true, "yclients.ru.evil.example": false, "fakeyclients.ru": false, "example.com": false,
	} {
		if got := bookingHost(host); got != want {
			t.Errorf("bookingHost(%q)=%v, want %v", host, got, want)
		}
	}
}
