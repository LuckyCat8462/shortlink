package stats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"shortlink/internal/channel"
	"shortlink/model"
	"shortlink/pkg/urltool"

	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/logx"
)

// Demo short codes used by seed (CHAR(6)). Stable for assistant demos.
const (
	DemoCodeSMS     = "demos1"
	DemoCodeAppPush = "demop1"
	DemoCodeWeChat  = "demow1"
)

// DemoShortCodes lists all seeded codes.
func DemoShortCodes() []string {
	return []string{DemoCodeSMS, DemoCodeAppPush, DemoCodeWeChat}
}

// ExpectedYesterdayTopChannel is guaranteed by SeedDemo ranking.
func ExpectedYesterdayTopChannel() string {
	return channel.SMS
}

type demoLink struct {
	sceneID   string
	shortCode string
	longURL   string
}

var demoLinks = []demoLink{
	{channel.SMS, DemoCodeSMS, "https://www.example.com/demo/sms"},
	{channel.AppPush, DemoCodeAppPush, "https://www.example.com/demo/app-push"},
	{channel.WeChat, DemoCodeWeChat, "https://www.example.com/demo/wechat"},
}

// SeedDemo upserts demo short links and rebuilds their click_events for the last 8 Shanghai days.
// Idempotent: deletes prior events for demo codes, then inserts a deterministic pattern.
// Yesterday (Shanghai): sms > app_push > wechat.
func SeedDemo(ctx context.Context, links model.ShortUrlMapModel, clicks model.ClickEventModel, filter *bloom.Filter) error {
	for _, dl := range demoLinks {
		if err := ensureDemoLink(ctx, links, dl); err != nil {
			return err
		}
		if filter != nil {
			if err := filter.Add([]byte(dl.shortCode)); err != nil {
				logx.Errorw("seed bloom add failed", logx.LogField{Key: "code", Value: dl.shortCode}, logx.LogField{Key: "err", Value: err.Error()})
			}
		}
	}

	if err := clicks.DeleteByShortCodes(ctx, DemoShortCodes()); err != nil {
		return fmt.Errorf("clear demo clicks: %w", err)
	}

	now := time.Now()
	loc := Location()
	y, m, d := now.In(loc).Date()
	today := time.Date(y, m, d, 12, 0, 0, 0, loc) // noon Shanghai for stable intra-day

	// pvByScene[dayOffset][scene] where dayOffset 0 = today, 1 = yesterday, ...
	// Yesterday (offset 1): sms 50, app_push 30, wechat 20
	type daySpec map[string]int
	specs := []daySpec{
		{channel.SMS: 12, channel.AppPush: 10, channel.WeChat: 8},  // today
		{channel.SMS: 50, channel.AppPush: 30, channel.WeChat: 20}, // yesterday — sms wins
		{channel.SMS: 18, channel.AppPush: 22, channel.WeChat: 15},
		{channel.SMS: 14, channel.AppPush: 16, channel.WeChat: 19},
		{channel.SMS: 25, channel.AppPush: 11, channel.WeChat: 9},
		{channel.SMS: 9, channel.AppPush: 21, channel.WeChat: 13},
		{channel.SMS: 17, channel.AppPush: 14, channel.WeChat: 12},
		{channel.SMS: 11, channel.AppPush: 8, channel.WeChat: 7}, // 8th day back
	}

	codeByScene := map[string]string{
		channel.SMS:     DemoCodeSMS,
		channel.AppPush: DemoCodeAppPush,
		channel.WeChat:  DemoCodeWeChat,
	}

	rows := make([]*model.ClickEvent, 0, 512)
	for offset, spec := range specs {
		day := today.AddDate(0, 0, -offset)
		for sceneID, n := range spec {
			code := codeByScene[sceneID]
			for i := 0; i < n; i++ {
				// spread within the day; unique visitors every 3 clicks
				clicked := day.Add(time.Duration(i) * time.Minute)
				visitor := fmt.Sprintf("seed|%s|%s|%d", sceneID, FormatDate(day), i/3)
				rows = append(rows, &model.ClickEvent{
					ShortCode:   code,
					SceneId:     sceneID,
					VisitorHash: VisitorHash(visitor, "seed-ua"),
					ClickedAt:   clicked.UTC(),
				})
			}
		}
	}

	if err := clicks.InsertBatch(ctx, rows); err != nil {
		return fmt.Errorf("insert demo clicks: %w", err)
	}
	logx.Infof("SeedDemo: links=%d click_events=%d yesterday_top=%s",
		len(demoLinks), len(rows), ExpectedYesterdayTopChannel())
	return nil
}

func ensureDemoLink(ctx context.Context, links model.ShortUrlMapModel, dl demoLink) error {
	if _, err := links.FindOneByShortCode(ctx, dl.shortCode); err == nil {
		return nil
	} else if !errors.Is(err, model.ErrNotFound) {
		return err
	}
	hash := urltool.URLHash(dl.sceneID, dl.longURL)
	_, err := links.Insert(ctx, &model.ShortUrlMap{
		SceneId:   dl.sceneID,
		UrlHash:   hash,
		LongUrl:   dl.longURL,
		ShortCode: dl.shortCode,
		Status:    1,
		ExpireAt:  sql.NullTime{},
		CreatedBy: "seed-demo",
	})
	if err != nil {
		if model.IsDuplicateErr(err) {
			return nil
		}
		return fmt.Errorf("insert demo link %s: %w", dl.shortCode, err)
	}
	return nil
}
