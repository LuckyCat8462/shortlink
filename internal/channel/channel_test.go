package channel

import "testing"

func TestDisplayName(t *testing.T) {
	if DisplayName(SMS) != "短信" {
		t.Fatal(DisplayName(SMS))
	}
	if DisplayName(AppPush) != "App 推送" {
		t.Fatal(DisplayName(AppPush))
	}
	if DisplayName(WeChat) != "微信" {
		t.Fatal(DisplayName(WeChat))
	}
	if DisplayName("") != "未分类" {
		t.Fatal(DisplayName(""))
	}
	if DisplayName("other") != "other" {
		t.Fatal(DisplayName("other"))
	}
}

func TestAllStable(t *testing.T) {
	all := All()
	if len(all) != 3 || all[0] != SMS || all[1] != AppPush || all[2] != WeChat {
		t.Fatalf("%v", all)
	}
}
