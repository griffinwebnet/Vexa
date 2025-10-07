package utils

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditContext holds audit information for a request
type AuditContext struct {
	User      string
	IPAddress string
	UserAgent string
	SessionID string
}

// GetAuditContext extracts audit information from Gin context
func GetAuditContext(c *gin.Context) AuditContext {
	// Get user from JWT claims if available
	user := "anonymous"
	if claims, exists := c.Get("user_claims"); exists {
		if userClaims, ok := claims.(map[string]interface{}); ok {
			if username, exists := userClaims["username"]; exists {
				if usernameStr, ok := username.(string); ok {
					user = usernameStr
				}
			}
		}
	}

	// Get IP address
	ip := c.ClientIP()
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
	}

	// Get user agent
	userAgent := c.GetHeader("User-Agent")

	// Get session ID from header or cookie
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		if cookie, err := c.Cookie("session_id"); err == nil {
			sessionID = cookie
		}
	}

	return AuditContext{
		User:      user,
		IPAddress: ip,
		UserAgent: userAgent,
		SessionID: sessionID,
	}
}

// LogAuthentication logs authentication events
func LogAuthentication(ctx AuditContext, action string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "authentication",
		Resource:  "system",
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogUserManagement logs user management events
func LogUserManagement(ctx AuditContext, action string, targetUser string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["target_user"] = targetUser

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "user_management",
		Resource:  "user:" + targetUser,
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogGroupManagement logs group management events
func LogGroupManagement(ctx AuditContext, action string, groupName string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["group_name"] = groupName

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "group_management",
		Resource:  "group:" + groupName,
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogDomainManagement logs domain management events
func LogDomainManagement(ctx AuditContext, action string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "domain_management",
		Resource:  "domain",
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogComputerManagement logs computer management events
func LogComputerManagement(ctx AuditContext, action string, computerName string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["computer_name"] = computerName

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "computer_management",
		Resource:  "computer:" + computerName,
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogSystemManagement logs system management events
func LogSystemManagement(ctx AuditContext, action string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "system_management",
		Resource:  "system",
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogDataAccess logs data access events
func LogDataAccess(ctx AuditContext, action string, resource string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "data_access",
		Resource:  resource,
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(ctx AuditContext, action string, severity string, success bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["severity"] = severity

	event := AuditEvent{
		Timestamp: time.Now(),
		User:      ctx.User,
		Action:    action,
		Category:  "security",
		Resource:  "system",
		Details:   details,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		SessionID: ctx.SessionID,
		Success:   success,
	}

	Audit(event)
}

// AuditMiddleware logs all API requests
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health checks and static assets
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/assets") {
			c.Next()
			return
		}

		ctx := GetAuditContext(c)

		// Log the request
		LogDataAccess(ctx, "api_request", path, true, map[string]interface{}{
			"method":  c.Request.Method,
			"path":    path,
			"query":   c.Request.URL.RawQuery,
			"referer": c.GetHeader("Referer"),
		})

		c.Next()

		// Log response status
		status := c.Writer.Status()
		success := status >= 200 && status < 400

		LogDataAccess(ctx, "api_response", path, success, map[string]interface{}{
			"method": c.Request.Method,
			"path":   path,
			"status": status,
			"size":   c.Writer.Size(),
		})
	}
}
