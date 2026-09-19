package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOW_ORIGINS"))

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" && isOriginAllowed(origin, c.Request, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			if origin != "" && !isOriginAllowed(origin, c.Request, allowedOrigins) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out[strings.ToLower(item)] = struct{}{}
	}
	return out
}

func isOriginAllowed(origin string, req *http.Request, allowList map[string]struct{}) bool {
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}
	originHost := strings.ToLower(originURL.Host)

	// Same-host origin is always allowed.
	if originHost == strings.ToLower(requestHost(req)) {
		return true
	}

	// Additional allow-list from env.
	if _, ok := allowList[originHost]; ok {
		return true
	}
	if _, ok := allowList[strings.ToLower(origin)]; ok {
		return true
	}
	return false
}

func requestHost(req *http.Request) string {
	if h := strings.TrimSpace(req.Header.Get("X-Forwarded-Host")); h != "" {
		return strings.Split(h, ",")[0]
	}
	return req.Host
}
