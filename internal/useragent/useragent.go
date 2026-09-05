package useragent

import (
	useragent "github.com/mssola/user_agent"
)

// Parse extracts browser, os and a coarse device class from a raw User-Agent
// header. Anything the parser can't determine comes back nil so the column
// stays null rather than storing a meaningless empty string.
func Parse(raw string) (browser *string, os *string, device *string) {
	if raw == "" {
		return nil, nil, nil
	}

	ua := useragent.New(raw)

	if name, _ := ua.Browser(); name != "" {
		browser = &name
	}

	if osName := ua.OS(); osName != "" {
		os = &osName
	}

	// mssola has no reliable tablet detection, so tablets classify as mobile.
	class := "desktop"
	switch {
	case ua.Bot():
		class = "bot"
	case ua.Mobile():
		class = "mobile"
	}
	device = &class

	return browser, os, device
}
