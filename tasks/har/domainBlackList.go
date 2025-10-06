package har

import "strings"

var thirdPartyBlocklist = []string{
	"google-analytics.com",
	"analytics.google.com",
	"googletagmanager.com",
	"googlesyndication.com",
	"fonts.googleapis.com",
	"fonts.gstatic.com",
	"doubleclick.net",
	"googleapis.com", // caution: might remove API calls to google; keep if we don't use Google
	"gstatic.com",
	"facebook.com",
	"connect.facebook.net",
	"ads.google.com",
	"scorecardresearch.com",
	"adsrvr.org",
	"adservice.google.com",
	"stripe.com",
	"checkout.stripe.com",
	"sentry.io",
	"hotjar.com",
	"mixpanel.com",
	"segment.com",
	"intercom.io",
	"newrelic.com",
	"cloudflare.com",
	"cdn.jsdelivr.net",
	"unpkg.com",
	"crashlytics.com",
}

func isBlockedDomain(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}

	for _, blocked := range thirdPartyBlocklist {
		blocked = strings.ToLower(strings.TrimSpace(blocked))
		if blocked == "" {
			continue
		}
		if host == blocked || strings.HasSuffix(host, "."+blocked) {
			return true
		}
	}

	return false
}
