package channel

// Demo channel scene_id values (hardcoded for assistant / seed).
const (
	SMS     = "sms"
	AppPush = "app_push"
	WeChat  = "wechat"
)

var displayNames = map[string]string{
	SMS:     "短信",
	AppPush: "App 推送",
	WeChat:  "微信",
}

// All returns demo channels in stable order.
func All() []string {
	return []string{SMS, AppPush, WeChat}
}

// DisplayName returns the Chinese label for a scene_id; unknown ids echo themselves.
func DisplayName(sceneID string) string {
	if name, ok := displayNames[sceneID]; ok {
		return name
	}
	if sceneID == "" {
		return "未分类"
	}
	return sceneID
}

// IsKnown reports whether sceneID is one of the demo channels.
func IsKnown(sceneID string) bool {
	_, ok := displayNames[sceneID]
	return ok
}
