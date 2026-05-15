package logic

import (
	"context"
	"database/sql"
	"errors"
	"shorturl/internal/config"
	"shorturl/internal/svc"
	"shorturl/internal/types"
	"shorturl/model"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ShortUrlMapModelInterface 定义测试所需的接口方法
type ShortUrlMapModelInterface interface {
	FindOneByMd5(ctx context.Context, md5Value sql.NullString) (*model.ShortUrlMap, error)
	FindOneBySurl(ctx context.Context, surl sql.NullString) (*model.ShortUrlMap, error)
	Insert(ctx context.Context, data *model.ShortUrlMap) (sql.Result, error)
}

// MockSequence 模拟发号器
type MockSequence struct {
	nextID uint64
}

func (m *MockSequence) Next() (uint64, error) {
	m.nextID++
	return m.nextID, nil
}

// MockShortUrlMapModel 模拟数据库模型
type MockShortUrlMapModel struct {
	data      map[string]*model.ShortUrlMap
	insertErr error
}

func NewMockShortUrlMapModel() *MockShortUrlMapModel {
	return &MockShortUrlMapModel{
		data: make(map[string]*model.ShortUrlMap),
	}
}

func (m *MockShortUrlMapModel) FindOne(ctx context.Context, id uint64) (*model.ShortUrlMap, error) {
	return nil, sqlx.ErrNotFound
}

func (m *MockShortUrlMapModel) FindOneByMd5(ctx context.Context, md5Value sql.NullString) (*model.ShortUrlMap, error) {
	for _, v := range m.data {
		if v.Md5 == md5Value {
			return v, nil
		}
	}
	return nil, sqlx.ErrNotFound
}

func (m *MockShortUrlMapModel) FindOneBySurl(ctx context.Context, surl sql.NullString) (*model.ShortUrlMap, error) {
	if v, ok := m.data[surl.String]; ok {
		return v, nil
	}
	return nil, sqlx.ErrNotFound
}

func (m *MockShortUrlMapModel) Insert(ctx context.Context, data *model.ShortUrlMap) (sql.Result, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	m.data[data.Surl.String] = data
	return &MockResult{}, nil
}

func (m *MockShortUrlMapModel) Update(ctx context.Context, data *model.ShortUrlMap) error {
	return nil
}

func (m *MockShortUrlMapModel) Delete(ctx context.Context, id uint64) error {
	return nil
}

func (m *MockShortUrlMapModel) withSession(session sqlx.Session) model.ShortUrlMapModel {
	return m
}

type MockResult struct{}

func (m *MockResult) LastInsertId() (int64, error) { return 1, nil }
func (m *MockResult) RowsAffected() (int64, error) { return 1, nil }

// MockServiceContext 实现 ServiceContext 的部分接口
type MockServiceContext struct {
	Config            config.Config
	Sequence          *MockSequence
	ShortUrlModel     *MockShortUrlMapModel
	ShortURLBlacklist map[string]struct{}
}

// createTestServiceContext 创建测试用的 ServiceContext
func createTestServiceContext(blacklist map[string]struct{}) *MockServiceContext {
	return &MockServiceContext{
		Config: config.Config{
			ShortDomain: "http://short.url",
			ShortURLBlacklist: func() []string {
				result := make([]string, 0, len(blacklist))
				for k := range blacklist {
					result = append(result, k)
				}
				return result
			}(),
		},
		Sequence:          &MockSequence{},
		ShortUrlModel:     NewMockShortUrlMapModel(),
		ShortURLBlacklist: blacklist,
	}
}

// getServiceContext 将 MockServiceContext 转换为 *svc.ServiceContext
func getServiceContext(m *MockServiceContext) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config:            m.Config,
		Sequence:          m.Sequence,
		ShortUrlModel:     m.ShortUrlModel,
		ShortURLBlacklist: m.ShortURLBlacklist,
	}
}

// TestShorturl_Success 测试成功转链
func TestShorturl_Success(t *testing.T) {
	mockSvc := createTestServiceContext(make(map[string]struct{}))

	logic := NewShorturlLogic(context.Background(), getServiceContext(mockSvc))

	req := &types.ConvertRequest{
		LongURL: "https://example.com/page",
	}

	resp, err := logic.Shorturl(req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.ShortURL == "" {
		t.Error("Expected non-empty ShortURL")
	}
	t.Logf("Successfully generated short URL: %s", resp.ShortURL)
}

// TestShorturl_AlreadyConverted 测试重复转链
func TestShorturl_AlreadyConverted(t *testing.T) {
	blacklist := make(map[string]struct{})
	mockSvc := createTestServiceContext(blacklist)

	existingURL := "https://example.com/page"
	mockSvc.ShortUrlModel.data["abc123"] = &model.ShortUrlMap{
		Lurl: sql.NullString{String: existingURL, Valid: true},
		Surl: sql.NullString{String: "abc123", Valid: true},
		Md5:  sql.NullString{String: "existing_md5", Valid: true},
	}

	logic := NewShorturlLogic(context.Background(), getServiceContext(mockSvc))

	req := &types.ConvertRequest{
		LongURL: existingURL,
	}

	_, err := logic.Shorturl(req)

	if err == nil {
		t.Error("Expected error for already converted URL, got nil")
	}
	t.Logf("Correctly rejected already converted URL: %v", err)
}

// TestShorturl_Blacklist 测试黑名单
func TestShorturl_Blacklist(t *testing.T) {
	blacklist := map[string]struct{}{"10": {}}
	mockSvc := createTestServiceContext(blacklist)

	// 重置 mock sequence 的 nextID
	mockSvc.Sequence.nextID = 61

	logic := NewShorturlLogic(context.Background(), getServiceContext(mockSvc))

	req := &types.ConvertRequest{
		LongURL: "https://example.com/page1",
	}

	_, err := logic.Shorturl(req)

	if err == nil {
		t.Error("Expected error for blacklisted URL, got nil")
	}
	t.Logf("Correctly rejected blacklisted short URL: %v", err)
}

// TestShorturl_InsertError 测试数据库插入错误
func TestShorturl_InsertError(t *testing.T) {
	blacklist := make(map[string]struct{})
	mockSvc := createTestServiceContext(blacklist)
	mockSvc.ShortUrlModel.insertErr = errors.New("database error")

	logic := NewShorturlLogic(context.Background(), getServiceContext(mockSvc))

	req := &types.ConvertRequest{
		LongURL: "https://example.com/page",
	}

	_, err := logic.Shorturl(req)

	if err == nil {
		t.Error("Expected error for insert failure, got nil")
	}
	t.Logf("Correctly handled insert error: %v", err)
}
