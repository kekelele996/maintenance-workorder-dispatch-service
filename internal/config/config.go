package config

import (
	"os"
	"strconv"
	"strings"
)

// Config 汇聚服务运行期配置，全部来自环境变量。
type Config struct {
	Workers        int
	RetryLimit     int
	ExecTimeoutMs  int
	PollIntervalMs int
	// SkillRoutes 记录设备类别到所需技能标签的映射，用于派单时匹配技工。
	SkillRoutes map[string]string
}

// defaultSkillRoutes 返回内置的设备类别 -> 技能标签默认路由表。
func defaultSkillRoutes() map[string]string {
	return map[string]string{
		"cnc":      "cnc-repair",
		"boiler":   "boiler-cert",
		"conveyor": "conveyor-maintenance",
		"hvac":     "hvac-cert",
		"default":  "general",
	}
}

// loadSkillRoutes 从环境变量加载自定义路由，未设置时回退到内置默认表。
func loadSkillRoutes() map[string]string {
	v := os.Getenv("WORKORDER_SKILL_ROUTES")
	if v == "" {
		return defaultSkillRoutes()
	}
	routes := map[string]string{}
	for _, pair := range strings.Split(v, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			routes[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(routes) == 0 {
		return nil
	}
	return routes
}

// Load 从环境变量加载配置；未设置时使用内置默认值。
func Load() *Config {
	return &Config{
		Workers:        getInt("WORKORDER_WORKERS", 4),
		RetryLimit:     getInt("WORKORDER_RETRY_LIMIT", 3),
		ExecTimeoutMs:  getInt("WORKORDER_EXEC_TIMEOUT_MS", 2000),
		PollIntervalMs: getInt("WORKORDER_POLL_INTERVAL_MS", 500),
		SkillRoutes:    loadSkillRoutes(),
	}
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func getList(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
