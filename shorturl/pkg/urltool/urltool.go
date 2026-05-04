package urltool

import (
	"net/url"
	"path"
)

func GetBasePath(targeturl string) (string, error) {
	myUrl, err := url.Parse(targeturl)
	if err != nil {
		return "", err
	}
	return path.Base(myUrl.Path), nil
}
