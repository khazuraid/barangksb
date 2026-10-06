package handlers

import "fmt"

func formatGeo(lat, lng float64) string {
	return fmt.Sprintf("%.6f, %.6f", lat, lng)
}
