// Package config 负责从环境变量加载并校验服务配置。
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultAdminUsername 是未显式配置时的管理员用户名。
	DefaultAdminUsername = "admin"
	// DefaultSiteName 是管理后台展示的站点名称。
	DefaultSiteName = "Feedbacker"
)

// Config 汇总服务运行所需的全部配置项。
type Config struct {
	Port               string
	MongoURI           string
	MongoDatabase      string
	AdminUsername      string
	JWTSecret          string
	JWTTTL             time.Duration
	CORSOrigins        []string
	TrustProxy         bool
	FeedbackAPIKey     string
	RateLimitPerMinute int
	SiteName           string
	LogLevel           slog.Level
	AdminDist          string
}

// Load 读取环境变量并返回校验后的配置。
func Load() (*Config, error) {
	cfg := &Config{
		Port:               env("PORT", "9090"),
		MongoURI:           env("MONGO_URI", "mongodb://127.0.0.1:27017"),
		MongoDatabase:      env("MONGO_DB", "feedbacker"),
		AdminUsername:      env("ADMIN_USERNAME", DefaultAdminUsername),
		JWTSecret:          env("JWT_SECRET", ""),
		CORSOrigins:        splitList(env("CORS_ORIGINS", "*")),
		TrustProxy:         envBool("TRUST_PROXY", false),
		FeedbackAPIKey:     env("FEEDBACK_API_KEY", ""),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 60),
		SiteName:           env("SITE_NAME", DefaultSiteName),
		AdminDist:          env("ADMIN_DIST", "../admin/dist"),
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("PORT 必须是 1-65535 之间的整数，当前值：%q", cfg.Port)
	}

	ttlHours := envInt("JWT_TTL_HOURS", 12)
	if ttlHours < 1 || ttlHours > 720 {
		return nil, fmt.Errorf("JWT_TTL_HOURS 必须在 1-720 之间，当前值：%d", ttlHours)
	}
	cfg.JWTTTL = time.Duration(ttlHours) * time.Hour

	if cfg.JWTSecret != "" && len(cfg.JWTSecret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET 长度至少为 16 个字符")
	}

	if cfg.RateLimitPerMinute < 1 || cfg.RateLimitPerMinute > 10000 {
		return nil, fmt.Errorf("RATE_LIMIT_PER_MINUTE 必须在 1-10000 之间，当前值：%d", cfg.RateLimitPerMinute)
	}

	if strings.TrimSpace(cfg.AdminUsername) == "" {
		return nil, fmt.Errorf("ADMIN_USERNAME 不能为空")
	}

	if len(cfg.CORSOrigins) == 0 {
		cfg.CORSOrigins = []string{"*"}
	}

	level := slog.LevelInfo
	if raw := strings.TrimSpace(os.Getenv("LOG_LEVEL")); raw != "" {
		if err := level.UnmarshalText([]byte(strings.ToUpper(raw))); err != nil {
			return nil, fmt.Errorf("LOG_LEVEL 只能是 debug/info/warn/error，当前值：%q", raw)
		}
	}
	cfg.LogLevel = level

	return cfg, nil
}

// DevMode 由构建标签决定，这里仅提供默认值，避免在常规构建中误用。
func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	v, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

func envInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
