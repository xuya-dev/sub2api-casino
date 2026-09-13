package casino

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/casino/gamemgr"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// fakeSettingStore 无 DB 的配置存储桩：gamemgr 首次加载会写入默认配置（此处仅吞掉）。
type fakeSettingStore struct{}

func (fakeSettingStore) GetSettingJSON(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (fakeSettingStore) PutSettingJSON(ctx context.Context, key, value string) error {
	return nil
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mgr, err := gamemgr.New(context.Background(), fakeSettingStore{})
	if err != nil {
		t.Fatalf("gamemgr.New: %v", err)
	}
	return &Server{mgr: mgr, limitPerSec: 5}
}

// injectAuth 把主站鉴权中间件注入的上下文键手动写进测试 context。
func injectAuth(c *gin.Context, userID int64, role string) {
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
	c.Set(string(middleware.ContextKeyUserRole), role)
}

// TestHandleMetaEnvelope 验证 /meta 走主站标准信封 {code:0,message,data}，
// 且身份取自中间件注入的 AuthSubject（无需自有会话）。
func TestHandleMetaEnvelope(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/casino/meta", nil)
	injectAuth(c, 42, "user")

	srv.HandleMeta(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want 200", rec.Code)
	}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			MinBet         float64 `json:"min_bet"`
			MaxBet         float64 `json:"max_bet"`
			DailyLossLimit float64 `json:"daily_loss_limit"`
			Wheel          struct {
				Segments []map[string]any `json:"segments"`
			} `json:"wheel"`
			Slots struct {
				Symbols []map[string]any `json:"symbols"`
			} `json:"slots"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应不是合法 JSON: %v\n%s", err, rec.Body.String())
	}
	if body.Code != 0 || body.Message != "success" {
		t.Fatalf("信封异常: code=%d message=%q", body.Code, body.Message)
	}
	if body.Data.MinBet <= 0 || body.Data.MaxBet < body.Data.MinBet {
		t.Fatalf("下注范围异常: min=%v max=%v", body.Data.MinBet, body.Data.MaxBet)
	}
	if len(body.Data.Wheel.Segments) == 0 || len(body.Data.Slots.Symbols) == 0 {
		t.Fatal("meta 缺少 wheel/symbols 展示信息")
	}
}

// TestRateLimitAllow 验证每用户滑动窗口限流：超过阈值后拒绝，其他用户不受影响。
func TestRateLimitAllow(t *testing.T) {
	srv := newTestServer(t)
	for i := 0; i < 5; i++ {
		if !srv.allow(1) {
			t.Fatalf("第 %d 次请求不应被限流", i+1)
		}
	}
	if srv.allow(1) {
		t.Fatal("同一用户第 6 次请求应被限流")
	}
	if !srv.allow(2) {
		t.Fatal("另一用户不应受影响")
	}
}

// TestCurrentUserIDsMissing 验证未注入 AuthSubject 时识别为未登录。
func TestCurrentUserIDsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if _, ok := currentUserID(c); ok {
		t.Fatal("无 AuthSubject 时不应取得用户 ID")
	}
}
