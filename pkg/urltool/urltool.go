package urltool

import (
	"errors"
	"net/url"
	"path"
	"strings"

	"shortlink/pkg/md5"
)

// NormalizeLongURL trims spaces and requires http/https. Does not alter query.
func NormalizeLongURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("longUrl is required")
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", errors.New("invalid longUrl")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("longUrl must start with http:// or https://")
	}
	if u.Host == "" {
		return "", errors.New("longUrl host is required")
	}
	return s, nil
}

// URLHash = MD5(scene_id + "\n" + normalized_long_url)
func URLHash(sceneID, normalizedLongURL string) string {
	return md5.Sum([]byte(sceneID + "\n" + normalizedLongURL))
}

// GetBasePath returns the last path segment (used to detect already-short URLs).
func GetBasePath(targeturl string) (string, error) {
	myUrl, err := url.Parse(targeturl)
	if err != nil {
		return "", err
	}
	if len(myUrl.Host) == 0 {
		return "", errors.New("target url is empty")
	}
	return path.Base(myUrl.Path), nil
}
