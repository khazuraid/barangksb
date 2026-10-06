package handlers

import (
	"testing"

	"inventariskantor/internal/service"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Peralatan Medis & Alkes (AKL/AKD)": "peralatan-medis-alkes-akl-akd",
		"ATK (Alat Tulis Kantor)":           "atk-alat-tulis-kantor",
		"Kebersihan  Sanitasi":              "kebersihan-sanitasi",
	}
	for in, want := range cases {
		if got := service.Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPager(t *testing.T) {
	p := viewsPager(t, 3, 25, 100)
	if p.Pages() != 4 {
		t.Errorf("Pages() = %d, want 4", p.Pages())
	}
	p2 := viewsPager(t, 1, 25, 5)
	if p2.Pages() != 1 {
		t.Errorf("Pages() = %d, want 1", p2.Pages())
	}
}

func TestRateLimitLogic(t *testing.T) {
	// limiter: 5 burst
	allowed := 0
	for i := 0; i < 10; i++ {
		if testLimiterAllow() {
			allowed++
		}
	}
	if allowed > 5 {
		t.Errorf("allowed %d > 5", allowed)
	}
}
