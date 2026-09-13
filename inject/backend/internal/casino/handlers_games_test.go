package casino

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// postJSON 构造带鉴权上下文的 POST 请求并调用 handler（不依赖存储层即可覆盖守卫与校验分支）。
func postJSON(t *testing.T, path, body string, h gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	injectAuth(c, 42, "user")
	h(c)
	return rec
}

// TestHandleSicboRollGuards 骰宝端点：娱乐模式关闭 403；注项非法 400（不触达存储层）。
func TestHandleSicboRollGuards(t *testing.T) {
	srv := newTestServer(t)

	disabled := false
	srv.mgr.GetConfig().Enabled = &disabled
	if rec := postJSON(t, "/games/sicbo/roll", `{"bet":1,"bet_type":"big"}`, srv.HandleSicboRoll); rec.Code != http.StatusForbidden {
		t.Fatalf("关闭娱乐模式应 403, got %d", rec.Code)
	}
	enabled := true
	srv.mgr.GetConfig().Enabled = &enabled

	if rec := postJSON(t, "/games/sicbo/roll", `{"bet":1,"bet_type":"huge"}`, srv.HandleSicboRoll); rec.Code != http.StatusBadRequest {
		t.Fatalf("非法注项应 400, got %d", rec.Code)
	}
	if rec := postJSON(t, "/games/sicbo/roll", `not-json`, srv.HandleSicboRoll); rec.Code != http.StatusBadRequest {
		t.Fatalf("格式错误应 400, got %d", rec.Code)
	}
}

// TestHandleBaccaratDealGuards 百家乐端点：娱乐模式关闭 403；方向非法 400（不触达存储层）。
func TestHandleBaccaratDealGuards(t *testing.T) {
	srv := newTestServer(t)

	disabled := false
	srv.mgr.GetConfig().Enabled = &disabled
	if rec := postJSON(t, "/games/baccarat/deal", `{"bet":1,"side":"player"}`, srv.HandleBaccaratDeal); rec.Code != http.StatusForbidden {
		t.Fatalf("关闭娱乐模式应 403, got %d", rec.Code)
	}
	enabled := true
	srv.mgr.GetConfig().Enabled = &enabled

	if rec := postJSON(t, "/games/baccarat/deal", `{"bet":1,"side":"dragon"}`, srv.HandleBaccaratDeal); rec.Code != http.StatusBadRequest {
		t.Fatalf("非法方向应 400, got %d", rec.Code)
	}
	if rec := postJSON(t, "/games/baccarat/deal", `not-json`, srv.HandleBaccaratDeal); rec.Code != http.StatusBadRequest {
		t.Fatalf("格式错误应 400, got %d", rec.Code)
	}
}

// TestHandleScratchRevealGuards 刮刮乐端点：娱乐模式关闭 403；
// 面值走 checkBet、数量须在 1~10（0/11 报 400），均不触达存储层。
func TestHandleScratchRevealGuards(t *testing.T) {
	srv := newTestServer(t)

	disabled := false
	srv.mgr.GetConfig().Enabled = &disabled
	if rec := postJSON(t, "/games/scratch/reveal", `{"face_value":1,"count":1}`, srv.HandleScratchReveal); rec.Code != http.StatusForbidden {
		t.Fatalf("关闭娱乐模式应 403, got %d", rec.Code)
	}
	enabled := true
	srv.mgr.GetConfig().Enabled = &enabled

	if rec := postJSON(t, "/games/scratch/reveal", `{"face_value":0,"count":1}`, srv.HandleScratchReveal); rec.Code != http.StatusBadRequest {
		t.Fatalf("面值为 0 应 400, got %d", rec.Code)
	}
	if rec := postJSON(t, "/games/scratch/reveal", `{"face_value":999,"count":1}`, srv.HandleScratchReveal); rec.Code != http.StatusBadRequest {
		t.Fatalf("面值超过最大下注应 400, got %d", rec.Code)
	}
	if rec := postJSON(t, "/games/scratch/reveal", `{"face_value":1,"count":0}`, srv.HandleScratchReveal); rec.Code != http.StatusBadRequest {
		t.Fatalf("数量为 0 应 400, got %d", rec.Code)
	}
	if rec := postJSON(t, "/games/scratch/reveal", `{"face_value":1,"count":11}`, srv.HandleScratchReveal); rec.Code != http.StatusBadRequest {
		t.Fatalf("数量为 11 应 400, got %d", rec.Code)
	}
	// 玩法校验：面值非法时三者都走不到存储层，用报错文案区分「通过玩法校验」与「未知玩法」
	if rec := postJSON(t, "/games/scratch/reveal", `{"face_value":0,"count":1,"mode":"xxx"}`, srv.HandleScratchReveal); rec.Code != http.StatusBadRequest || strings.Contains(rec.Body.String(), "未知玩法") == false {
		t.Fatalf("未知玩法应报未知玩法, got %d %s", rec.Code, rec.Body.String())
	}
	for _, mode := range []string{"classic", "lucky7", "lines"} {
		rec := postJSON(t, "/games/scratch/reveal", `{"face_value":0,"count":1,"mode":"`+mode+`"}`, srv.HandleScratchReveal)
		if rec.Code != http.StatusBadRequest || strings.Contains(rec.Body.String(), "下注金额") == false {
			t.Fatalf("合法玩法 %s 应通过玩法校验进入下注校验, got %d %s", mode, rec.Code, rec.Body.String())
		}
	}
	if rec := postJSON(t, "/games/scratch/reveal", `not-json`, srv.HandleScratchReveal); rec.Code != http.StatusBadRequest {
		t.Fatalf("格式错误应 400, got %d", rec.Code)
	}
}
