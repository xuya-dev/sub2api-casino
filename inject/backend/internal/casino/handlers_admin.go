package casino

import (
	"github.com/Wei-Shaw/sub2api/internal/casino/gamemgr"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// HandleGetConfig 管理端读取完整游戏配置 + 理论回报率预览。
func (s *Server) HandleGetConfig(c *gin.Context) {
	cfg := s.mgr.Get()
	response.Success(c, gin.H{"config": cfg, "rtp": cfg.RTP()})
}

// HandlePutConfig 管理端保存配置：校验 → 写库 → 热生效（无需重启）。
func (s *Server) HandlePutConfig(c *gin.Context) {
	var cfg gamemgr.Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "配置格式错误: "+err.Error())
		return
	}
	if err := s.mgr.Save(c.Request.Context(), &cfg); err != nil {
		response.BadRequest(c, "配置未通过校验: "+err.Error())
		return
	}
	saved := s.mgr.Get()
	response.Success(c, gin.H{"config": saved, "rtp": saved.RTP()})
}

// HandleAdminStats 管理看板：总盈亏 + 分游戏统计 + 最近下注。
func (s *Server) HandleAdminStats(c *gin.Context) {
	stats, err := s.st.AdminStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, "统计查询失败")
		return
	}
	gameStats, _ := s.st.GameStats(c.Request.Context())
	recent, _ := s.st.AdminRecentBets(c.Request.Context(), 20)
	if gameStats == nil {
		gameStats = []map[string]any{}
	}
	if recent == nil {
		recent = []map[string]any{}
	}
	response.Success(c, gin.H{"summary": stats, "games": gameStats, "recent": recent})
}
