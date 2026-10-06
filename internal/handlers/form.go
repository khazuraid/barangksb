package handlers

import "net/url"

// formValues lets handlers accept both url.Values (posts) and test fakes.
type formValues interface{ Get(key string) string }

var _ formValues = (*url.Values)(nil)
