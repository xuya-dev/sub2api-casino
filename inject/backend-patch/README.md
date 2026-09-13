# backend-patch：主站挂载补丁说明

本目录保存 casino 注入主站后端所需的"路由挂载补丁"。`inject.py` 注入时按以下规则操作（幂等）。

## 1. 新增文件

把本目录的 `casino.go` 拷贝到：

```
sub2api/backend/internal/server/routes/casino.go
```

内容为 routes 包风格的薄封装，直接转调 `internal/casino.RegisterCasinoRoutes`。
（真实树中的同名文件与本文件保持一致。）

## 2. router.go 标记块

文件：`sub2api/backend/internal/server/router.go`

位置：`registerRoutes` 函数体内、`handler.RegisterPageRoutes(...)` 之后（其它 Register 之后，函数收尾 `}` 之前）。
标记：`// ===CASINO:BEGIN===` 与 `// ===CASINO:END===` 之间整块替换为：

```go
	// ===CASINO:BEGIN===
	if err := routes.RegisterCasinoRoutes(v1, jwtAuth, adminAuth, cfg, redisClient); err != nil {
		log.Printf("casino routes register failed: %v", err)
	}
	// ===CASINO:END===
```

说明：

- 变量 `v1`、`jwtAuth`、`adminAuth`、`cfg`、`redisClient` 均为 `registerRoutes` 现有参数/局部变量，无需改签名。
- `log` 已在 router.go 导入清单中。
- `adminAuth`（`middleware.AdminAuthMiddleware`）自带 JWT 解析与管理员角色校验；`jwtAuth`（`middleware.JWTAuthMiddleware`）是命名类型，routes 层使用时按需 `gin.HandlerFunc(jwtAuth)` 转换。

## 3. 依赖

`sub2api/backend` 下需要 pgx 直接依赖（go.sum 已含 v5.7.4，可离线执行）：

```
go get github.com/jackc/pgx/v5@v5.7.4
```
