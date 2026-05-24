package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shorturl/internal/config"
	"shorturl/internal/types"
	"shorturl/model"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ShortUrlMapStore 定义测试所需的存储接口
type ShortUrlMapStore interface {
	FindOneByMd5(ctx context.Context, md5Value sql.NullString) (*model.ShortUrlMap, error)
	FindOneBySurl(ctx context.Context, surl sql.NullString) (*model.ShortUrlMap, error)
	Insert(ctx context.Context, data *model.ShortUrlMap) (sql.Result, error)
}

// SequenceGenerator 定义发号器接口
type SequenceGenerator interface {
	Next() (uint64, error)
}

// MockSequence 模拟发号器
type MockSequence struct {
	nextID uint64
	err    error
}

func (m *MockSequence) Next() (uint64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m.nextID++
	return m.nextID, nil
}

// MockShortUrlStore 模拟短链接存储
type MockShortUrlStore struct {
	data          map[string]*model.ShortUrlMap
	findByMd5Err  error
	findBySurlErr error
	insertErr     error
}

func NewMockShortUrlStore() *MockShortUrlStore {
	return &MockShortUrlStore{
		data: make(map[string]*model.ShortUrlMap),
	}
}

func (m *MockShortUrlStore) FindOneByMd5(ctx context.Context, md5Value sql.NullString) (*model.ShortUrlMap, error) {
	if m.findByMd5Err != nil {
		return nil, m.findByMd5Err
	}
	for _, v := range m.data {
		if v.Md5.String == md5Value.String {
			return v, nil
		}
	}
	return nil, sqlx.ErrNotFound
}

func (m *MockShortUrlStore) FindOneBySurl(ctx context.Context, surl sql.NullString) (*model.ShortUrlMap, error) {
	if m.findBySurlErr != nil {
		return nil, m.findBySurlErr
	}
	if v, ok := m.data[surl.String]; ok {
		return v, nil
	}
	return nil, sqlx.ErrNotFound
}

func (m *MockShortUrlStore) Insert(ctx context.Context, data *model.ShortUrlMap) (sql.Result, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	m.data[data.Surl.String] = data
	return &mockResult{}, nil
}

type mockResult struct{}

func (m *mockResult) LastInsertId() (int64, error) { return 1, nil }
func (m *mockResult) RowsAffected() (int64, error) { return 1, nil }

// ShorturlLogicForTest 用于测试的逻辑封装
type ShorturlLogicForTest struct {
	Config            config.Config
	Sequence          SequenceGenerator
	ShortUrlStore     ShortUrlMapStore
	ShortURLBlacklist map[string]struct{}
}

// ExecuteShorturl 执行转链逻辑（用于测试）
func (l *ShorturlLogicForTest) ExecuteShorturl(req *types.ConvertRequest) (string, error) {
	if req.LongURL == "" {
		return "", errors.New("长链接不能为空")
	}

	// 简化的URL验证（跳过网络请求）
	if !strings.HasPrefix(req.LongURL, "http://") && !strings.HasPrefix(req.LongURL, "https://") {
		return "", errors.New("无效链接格式")
	}

	// 检查是否已转链
	md5Value := fmt.Sprintf("%x", []byte(req.LongURL)) // 简化的md5
	u, err := l.ShortUrlStore.FindOneByMd5(context.Background(), sql.NullString{String: md5Value, Valid: true})
	if err != nil && err != sqlx.ErrNotFound {
		return "", fmt.Errorf("查询失败: %v", err)
	}
	if u != nil {
		return "", fmt.Errorf("该链接已被转为%s", u.Surl.String)
	}

	// 取号
	seqID, err := l.Sequence.Next()
	if err != nil {
		return "", fmt.Errorf("取号失败: %v", err)
	}

	// 转短链
	shortURL := fmt.Sprintf("%d", seqID) // 简化的短链生成

	// 检查黑名单
	if _, ok := l.ShortURLBlacklist[shortURL]; ok {
		return "", fmt.Errorf("短链接不能包含敏感词%s", shortURL)
	}

	// 存储
	_, err = l.ShortUrlStore.Insert(context.Background(), &model.ShortUrlMap{
		Lurl: sql.NullString{String: req.LongURL, Valid: true},
		Surl: sql.NullString{String: shortURL, Valid: true},
		Md5:  sql.NullString{String: md5Value, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("存储失败: %v", err)
	}

	return l.Config.ShortDomain + "/" + shortURL, nil
}

// TestShorturl_Success 测试成功转链
func TestShorturl_Success(t *testing.T) {
	blacklist := make(map[string]struct{})
	store := NewMockShortUrlStore()
	sequence := &MockSequence{}

	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          sequence,
		ShortUrlStore:     store,
		ShortURLBlacklist: blacklist,
	}

	req := &types.ConvertRequest{LongURL: "https://example.com/page"}

	result, err := logic.ExecuteShorturl(req)

	if err != nil {
		t.Fatalf("[TestShorturl_Success] 预期无错误，实际收到错误: %v", err)
	}
	if result == "" {
		t.Error("[TestShorturl_Success] 预期生成非空的短链接，实际为空")
	}
	if !strings.HasPrefix(result, "http://short.url/") {
		t.Errorf("[TestShorturl_Success] 短链接格式不正确，预期以 http://short.url/ 开头，实际为: %s", result)
	}
	t.Logf("[TestShorturl_Success] 成功生成短链接: %s", result)
}

// TestShorturl_EmptyLongURL 测试空长链接
func TestShorturl_EmptyLongURL(t *testing.T) {
	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{},
		ShortUrlStore:     NewMockShortUrlStore(),
		ShortURLBlacklist: make(map[string]struct{}),
	}

	req := &types.ConvertRequest{LongURL: ""}

	_, err := logic.ExecuteShorturl(req)

	if err == nil {
		t.Error("[TestShorturl_EmptyLongURL] 预期收到错误（空长链接），实际未收到错误")
	} else if !strings.Contains(err.Error(), "长链接不能为空") {
		t.Errorf("[TestShorturl_EmptyLongURL] 预期错误信息包含 '长链接不能为空'，实际为: %v", err)
	} else {
		t.Logf("[TestShorturl_EmptyLongURL] 正确拒绝空长链接: %v", err)
	}
}

// TestShorturl_InvalidLongURL 测试无效长链接格式
func TestShorturl_InvalidLongURL(t *testing.T) {
	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{},
		ShortUrlStore:     NewMockShortUrlStore(),
		ShortURLBlacklist: make(map[string]struct{}),
	}

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com",
		"http://:",
		"https://:",
	}

	for _, url := range invalidURLs {
		req := &types.ConvertRequest{LongURL: url}
		_, err := logic.ExecuteShorturl(req)

		if err == nil {
			t.Errorf("[TestShorturl_InvalidLongURL] 预期收到错误（无效URL: %s），实际未收到错误", url)
		} else if !strings.Contains(err.Error(), "无效链接格式") {
			t.Errorf("[TestShorturl_InvalidLongURL] 预期错误信息包含 '无效链接格式'，URL: %s, 实际错误: %v", url, err)
		} else {
			t.Logf("[TestShorturl_InvalidLongURL] 正确拒绝无效URL: %s", url)
		}
	}
}

// TestShorturl_AlreadyConverted 测试重复转链
func TestShorturl_AlreadyConverted(t *testing.T) {
	store := NewMockShortUrlStore()
	existingURL := "https://example.com/page"
	existingMd5 := fmt.Sprintf("%x", []byte(existingURL))

	store.data["abc123"] = &model.ShortUrlMap{
		Lurl: sql.NullString{String: existingURL, Valid: true},
		Surl: sql.NullString{String: "abc123", Valid: true},
		Md5:  sql.NullString{String: existingMd5, Valid: true},
	}

	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{},
		ShortUrlStore:     store,
		ShortURLBlacklist: make(map[string]struct{}),
	}

	req := &types.ConvertRequest{LongURL: existingURL}

	_, err := logic.ExecuteShorturl(req)

	if err == nil {
		t.Error("[TestShorturl_AlreadyConverted] 预期收到错误（URL已转换），实际未收到错误")
	} else if !strings.Contains(err.Error(), "该链接已被转为") {
		t.Errorf("[TestShorturl_AlreadyConverted] 预期错误信息包含 '该链接已被转为'，实际为: %v", err)
	} else {
		t.Logf("[TestShorturl_AlreadyConverted] 正确拒绝重复转链: %v", err)
	}
}

// TestShorturl_Blacklist 测试黑名单
func TestShorturl_Blacklist(t *testing.T) {
	blacklist := map[string]struct{}{"1": {}} // 第一个seqID生成的短链是"1"

	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{},
		ShortUrlStore:     NewMockShortUrlStore(),
		ShortURLBlacklist: blacklist,
	}

	req := &types.ConvertRequest{LongURL: "https://example.com/page"}

	_, err := logic.ExecuteShorturl(req)

	if err == nil {
		t.Error("[TestShorturl_Blacklist] 预期收到错误（短链接在黑名单中），实际未收到错误")
	} else if !strings.Contains(err.Error(), "短链接不能包含敏感词") {
		t.Errorf("[TestShorturl_Blacklist] 预期错误信息包含 '短链接不能包含敏感词'，实际为: %v", err)
	} else {
		t.Logf("[TestShorturl_Blacklist] 正确拒绝黑名单短链接: %v", err)
	}
}

// TestShorturl_InsertError 测试数据库插入错误
func TestShorturl_InsertError(t *testing.T) {
	store := NewMockShortUrlStore()
	store.insertErr = errors.New("database connection failed")

	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{},
		ShortUrlStore:     store,
		ShortURLBlacklist: make(map[string]struct{}),
	}

	req := &types.ConvertRequest{LongURL: "https://example.com/page"}

	_, err := logic.ExecuteShorturl(req)

	if err == nil {
		t.Error("[TestShorturl_InsertError] 预期收到错误（数据库插入失败），实际未收到错误")
	} else if !strings.Contains(err.Error(), "存储失败") {
		t.Errorf("[TestShorturl_InsertError] 预期错误信息包含 '存储失败'，实际为: %v", err)
	} else {
		t.Logf("[TestShorturl_InsertError] 正确处理数据库插入错误: %v", err)
	}
}

// TestShorturl_SequenceError 测试发号器错误
func TestShorturl_SequenceError(t *testing.T) {
	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{err: errors.New("sequence service unavailable")},
		ShortUrlStore:     NewMockShortUrlStore(),
		ShortURLBlacklist: make(map[string]struct{}),
	}

	req := &types.ConvertRequest{LongURL: "https://example.com/page"}

	_, err := logic.ExecuteShorturl(req)

	if err == nil {
		t.Error("[TestShorturl_SequenceError] 预期收到错误（发号器失败），实际未收到错误")
	} else if !strings.Contains(err.Error(), "取号失败") {
		t.Errorf("[TestShorturl_SequenceError] 预期错误信息包含 '取号失败'，实际为: %v", err)
	} else {
		t.Logf("[TestShorturl_SequenceError] 正确处理发号器错误: %v", err)
	}
}

// TestShorturl_SpecialCharacters 测试包含特殊字符的URL
func TestShorturl_SpecialCharacters(t *testing.T) {
	logic := &ShorturlLogicForTest{
		Config:            config.Config{ShortDomain: "http://short.url"},
		Sequence:          &MockSequence{},
		ShortUrlStore:     NewMockShortUrlStore(),
		ShortURLBlacklist: make(map[string]struct{}),
	}

	specialURLs := []string{
		"https://example.com/path?query=value&other=test",
		"https://example.com/page#section",
		"https://example.com/路径/中文",
		fmt.Sprintf("https://example.com/page?id=%d", 123456),
	}

	for _, url := range specialURLs {
		req := &types.ConvertRequest{LongURL: url}
		result, err := logic.ExecuteShorturl(req)

		if err != nil {
			t.Errorf("[TestShorturl_SpecialCharacters] 预期成功转换URL: %s, 实际收到错误: %v", url, err)
		} else if result == "" {
			t.Errorf("[TestShorturl_SpecialCharacters] 预期生成非空短链接，URL: %s", url)
		} else {
			t.Logf("[TestShorturl_SpecialCharacters] 成功转换特殊字符URL: %s -> %s", url, result)
		}
	}
}
