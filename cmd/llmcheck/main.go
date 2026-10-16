package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"shortlink/internal/assistant"

	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cm, err := assistant.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cm.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: "你是短链分析助手。只做简短中文回复。"},
		{Role: schema.User, Content: "用一句话说明你会做什么（还不要查数据）。"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content)
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		fmt.Printf("tokens: %+v\n", resp.ResponseMeta.Usage)
	}
	_ = os.Stdout
}