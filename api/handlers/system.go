package handlers

import (
	"net/http"
	"os/exec"
	"runtime"

	"github.com/gin-gonic/gin"
)

type SystemStatusResponse struct {
	OS            string `json:"os"`
	Architecture  string `json:"architecture"`
	SambaInstalled bool  `json:"samba_installed"`
	BindInstalled bool   `json:"bind_installed"`
	APIVersion    string `json:"api_version"`
}

// SystemStatus returns system information
func SystemStatus(c *gin.Context) {
	// Check if samba-tool is available
	sambaCmd := exec.Command("which", "samba-tool")
	sambaInstalled := sambaCmd.Run() == nil

	// Check if BIND is available
	bindCmd := exec.Command("which", "named")
	bindInstalled := bindCmd.Run() == nil

	c.JSON(http.StatusOK, SystemStatusResponse{
		OS:            runtime.GOOS,
		Architecture:  runtime.GOARCH,
		SambaInstalled: sambaInstalled,
		BindInstalled:  bindInstalled,
		APIVersion:    "1.0.0",
	})
}

