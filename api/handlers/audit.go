package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/griffinwebnet/vexa/api/utils"
)

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Category  string                 `json:"category"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
}

// AuditLogResponse represents the response for audit log queries
type AuditLogResponse struct {
	Entries []AuditLogEntry `json:"entries"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	HasMore bool            `json:"has_more"`
}

// LogLevelRequest represents a request to change log level
type LogLevelRequest struct {
	Level string `json:"level" binding:"required,oneof=DEBUG INFO WARN ERROR"`
}

// GetAuditLogs returns audit log entries with filtering and pagination
func GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	user := c.Query("user")
	category := c.Query("category")
	action := c.Query("action")
	success := c.Query("success")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	// Get audit log file path
	auditLogPath := "/var/log/vexa/audit.log"

	// Check if file exists
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, AuditLogResponse{
			Entries: []AuditLogEntry{},
			Total:   0,
			Page:    page,
			Limit:   limit,
			HasMore: false,
		})
		return
	}

	// Read and parse audit logs
	entries, err := readAuditLogs(auditLogPath, page, limit, user, category, action, success, startDate, endDate)
	if err != nil {
		utils.Error("Failed to read audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read audit logs",
		})
		return
	}

	// Get total count for pagination
	total, err := countAuditLogs(auditLogPath, user, category, action, success, startDate, endDate)
	if err != nil {
		utils.Error("Failed to count audit logs: %v", err)
		total = len(entries)
	}

	hasMore := (page * limit) < total

	response := AuditLogResponse{
		Entries: entries,
		Total:   total,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemLogs returns system log entries (debug, info, warn, error)
func GetSystemLogs(c *gin.Context) {
	logType := c.Param("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Validate log type
	validTypes := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validTypes[logType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log type. Must be: debug, info, warn, or error",
		})
		return
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 100
	}

	// Get log file path
	logPath := fmt.Sprintf("/var/log/vexa/%s.log", logType)

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"entries":  []string{},
			"total":    0,
			"page":     page,
			"limit":    limit,
			"has_more": false,
		})
		return
	}

	// Read log entries
	entries, total, err := readSystemLogs(logPath, page, limit)
	if err != nil {
		utils.Error("Failed to read system logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read system logs",
		})
		return
	}

	hasMore := (page * limit) < total

	c.JSON(http.StatusOK, gin.H{
		"entries":  entries,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": hasMore,
	})
}

// GetLogStats returns statistics about the audit logs
func GetLogStats(c *gin.Context) {
	auditLogPath := "/var/log/vexa/audit.log"

	// Check if file exists
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"total_entries":     0,
			"entries_today":     0,
			"failed_logins":     0,
			"successful_logins": 0,
			"user_actions":      0,
			"system_actions":    0,
			"categories":        map[string]int{},
		})
		return
	}

	stats, err := calculateLogStats(auditLogPath)
	if err != nil {
		utils.Error("Failed to calculate log stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate log statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetLogLevel changes the system log level
func SetLogLevel(c *gin.Context) {
	var req LogLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Log the log level change
	ctx := utils.GetAuditContext(c)
	utils.LogSystemManagement(ctx, "log_level_change", true, map[string]interface{}{
		"new_level":      req.Level,
		"previous_level": "INFO", // TODO: Get actual current level
	})

	// TODO: Implement actual log level change
	// This would require modifying the logger configuration

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Log level changed to %s", req.Level),
		"level":   req.Level,
	})
}

// Helper functions

func readAuditLogs(path string, page, limit int, user, category, action, success, startDate, endDate string) ([]AuditLogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []AuditLogEntry
	scanner := bufio.NewScanner(file)

	// Read all lines first (for simplicity, in production you'd want to optimize this)
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// Reverse the lines to get newest first
	for i := len(allLines) - 1; i >= 0; i-- {
		line := allLines[i]
		if line == "" {
			continue
		}

		var entry AuditLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}

		// Apply filters
		if !matchesFilters(entry, user, category, action, success, startDate, endDate) {
			continue
		}

		entries = append(entries, entry)
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= len(entries) {
		return []AuditLogEntry{}, nil
	}

	if end > len(entries) {
		end = len(entries)
	}

	return entries[start:end], nil
}

func countAuditLogs(path string, user, category, action, success, startDate, endDate string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry AuditLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if matchesFilters(entry, user, category, action, success, startDate, endDate) {
			count++
		}
	}

	return count, nil
}

func matchesFilters(entry AuditLogEntry, user, category, action, success, startDate, endDate string) bool {
	if user != "" && !strings.Contains(strings.ToLower(entry.User), strings.ToLower(user)) {
		return false
	}

	if category != "" && entry.Category != category {
		return false
	}

	if action != "" && !strings.Contains(strings.ToLower(entry.Action), strings.ToLower(action)) {
		return false
	}

	if success != "" {
		successBool, err := strconv.ParseBool(success)
		if err == nil && entry.Success != successBool {
			return false
		}
	}

	if startDate != "" {
		if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
			if entry.Timestamp.Before(startTime) {
				return false
			}
		}
	}

	if endDate != "" {
		if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
			if entry.Timestamp.After(endTime.Add(24 * time.Hour)) {
				return false
			}
		}
	}

	return true
}

func readSystemLogs(path string, page, limit int) ([]string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Reverse to get newest first
	for i := len(lines) - 1; i >= 0; i-- {
		lines = append(lines, lines[i])
	}
	lines = lines[len(lines)/2:]

	total := len(lines)

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= len(lines) {
		return []string{}, total, nil
	}

	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end], total, nil
}

func calculateLogStats(path string) (map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := map[string]interface{}{
		"total_entries":     0,
		"entries_today":     0,
		"failed_logins":     0,
		"successful_logins": 0,
		"user_actions":      0,
		"system_actions":    0,
		"categories":        map[string]int{},
	}

	today := time.Now().Format("2006-01-02")
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry AuditLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		stats["total_entries"] = stats["total_entries"].(int) + 1

		// Count today's entries
		if entry.Timestamp.Format("2006-01-02") == today {
			stats["entries_today"] = stats["entries_today"].(int) + 1
		}

		// Count login attempts
		switch entry.Action {
		case "login_success":
			stats["successful_logins"] = stats["successful_logins"].(int) + 1
		case "login_failed":
			stats["failed_logins"] = stats["failed_logins"].(int) + 1
		}

		// Count by category
		switch entry.Category {
		case "user_management", "group_management":
			stats["user_actions"] = stats["user_actions"].(int) + 1
		case "system_management":
			stats["system_actions"] = stats["system_actions"].(int) + 1
		}

		// Count by category
		categories := stats["categories"].(map[string]int)
		categories[entry.Category]++
	}

	return stats, nil
}
