package urltool

import "testing"

func TestGetBasePath(t *testing.T) {
	type args struct {
		targeturl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{

		{
			name:    "基础实例01",
			args:    args{targeturl: "https://baidu.com/posts/go/unit-test-0/"},
			want:    "unit-test-0",
			wantErr: false,
		},
		{
			name:    "路径仅包含斜杠",
			args:    args{targeturl: "https://example.com/"},
			want:    "/",
			wantErr: false,
		},
		{
			name:    "无效路径",
			args:    args{targeturl: "xxxx/1123"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "空字符串",
			args:    args{targeturl: ""},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBasePath(tt.args.targeturl)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBasePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBasePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
