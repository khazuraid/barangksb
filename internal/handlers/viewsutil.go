package handlers

import (
	"os"

	"inventariskantor/internal/views"
)

// views helpers — keep handler bodies readable
type viewsItemRow = views.ItemRow

type viewsBarcodeData = views.BarcodeData

func viewsBarcodePage(d views.BarcodeData) interface{ Render() } { return nil }

func viewsDriveSyncData(d views.DriveSyncData) views.DriveSyncData { return d }

func osGetenv(k string) string { return os.Getenv(k) }

func osReadFileEnv(k string) ([]byte, error) { return os.ReadFile(os.Getenv(k)) }
