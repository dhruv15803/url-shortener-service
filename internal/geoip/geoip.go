package geoip

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

// Reader looks up geo information for an ip against a local MaxMind database.
// A Reader built from an empty path is "disabled" - every lookup returns nils
// instead of failing, so the service runs fine without the mmdb present.
type Reader struct {
	db *geoip2.Reader
}

func NewReader(path string) (*Reader, error) {
	if path == "" {
		return &Reader{}, nil
	}

	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}

	return &Reader{db: db}, nil
}

func (r *Reader) Enabled() bool {
	return r.db != nil
}

func (r *Reader) Close() error {
	if r.db == nil {
		return nil
	}
	return r.db.Close()
}

// Lookup returns country/city/region for an ip. Everything is nil when the
// reader is disabled, the ip is unparseable, or the address is private or
// loopback (which is why local testing shows no geo unless the ip is spoofed).
func (r *Reader) Lookup(ip string) (country *string, city *string, region *string) {
	if r.db == nil || ip == "" {
		return nil, nil, nil
	}

	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return nil, nil, nil
	}

	record, err := r.db.City(parsed)
	if err != nil {
		return nil, nil, nil
	}

	if record.Country.IsoCode != "" {
		code := record.Country.IsoCode
		country = &code
	}

	if name, ok := record.City.Names["en"]; ok && name != "" {
		city = &name
	}

	if len(record.Subdivisions) > 0 {
		if name, ok := record.Subdivisions[0].Names["en"]; ok && name != "" {
			region = &name
		}
	}

	return country, city, region
}
