package connect

import (
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"time"
)

// client是全局的HTTP客户端
var client = &http.Client{Transport: &http.Transport{DisableKeepAlives: true}, Timeout: 2 * time.Second}

// Get 转链时,判断链接是否是请求通
func Get(url string) bool {
	resp, err := client.Get(url)
	if err != nil {
		logx.Errorw("", logx.LogField{Key: "err", Value: err.Error()})
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == http.StatusOK // 别人发送的跳转响应，此处不算通过
}
