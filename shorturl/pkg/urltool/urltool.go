package urltool

import (
	"errors"
	"net/url"
	"path"
)

// GetBasePath 避免循环转链（输入不能为短链接）
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
