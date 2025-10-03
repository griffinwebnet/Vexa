package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vexa/api/services"
	"github.com/vexa/api/version"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int    `json:"size"`
	} `json:"assets"`
}

// SystemVersion represents version information for a system component
type SystemVersion struct {
	Component       string `json:"component"`        // Component name (api, web, samba, etc)
	Version         string `json:"version"`          // Current version
	UpdateStatus    string `json:"update_status"`    // Status of any pending updates
	LastUpdated     string `json:"last_updated"`     // When component was last updated
	Dependencies    bool   `json:"dependencies"`     // Whether component has all required dependencies
	RequiresRestart bool   `json:"requires_restart"` // Whether update requires service restart
}

// UpdateStatus represents the current update status
type UpdateStatus struct {
	Versions        []SystemVersion `json:"versions"`          // Current versions of all components
	LatestVersion   string          `json:"latest_version"`    // Latest version from GitHub
	UpdateAvailable bool            `json:"update_available"`  // Whether an update is available
	Status          string          `json:"status"`            // Status message (Up to Date, Update Available, etc)
	LatestRelease   *GitHubRelease  `json:"latest_release"`    // Details about the latest release
	Error           string          `json:"error,omitempty"`   // Any error that occurred
	Channel         string          `json:"channel"`           // Current update channel (stable/nightly)
	BuildFromSource bool            `json:"build_from_source"` // Whether updates build from source
}

// CheckForUpdates checks for available updates from GitHub releases
func CheckForUpdates(c *gin.Context) {
	// Get current versions
	componentVersions := version.Components()
	fmt.Printf("Component versions: %+v\n", componentVersions)

	// Get status of all managed components
	versions := []SystemVersion{
		{
			Component:       "api",
			Version:         version.Current,
			UpdateStatus:    "up_to_date",
			LastUpdated:     time.Now().Format(time.RFC3339),
			Dependencies:    true,
			RequiresRestart: true,
		},
		{
			Component:       "web",
			Version:         version.Current,
			UpdateStatus:    "up_to_date",
			LastUpdated:     time.Now().Format(time.RFC3339),
			Dependencies:    true,
			RequiresRestart: true,
		},
		{
			Component:       "samba",
			Version:         services.GetSambaVersion(),
			UpdateStatus:    services.CheckSambaUpdates(),
			LastUpdated:     services.GetSambaLastUpdate(),
			Dependencies:    services.CheckSambaDependencies(),
			RequiresRestart: true,
		},
		{
			Component:       "headscale",
			Version:         services.GetHeadscaleVersion(),
			UpdateStatus:    services.CheckHeadscaleUpdates(),
			LastUpdated:     services.GetHeadscaleLastUpdate(),
			Dependencies:    services.CheckHeadscaleDependencies(),
			RequiresRestart: true,
		},
	}

	// Get repository info
	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		repo = "griffinwebnet/Vexa"
	}

	// Fetch releases from GitHub
	releases, err := fetchGitHubReleases(repo)
	if err != nil {
		c.JSON(http.StatusOK, UpdateStatus{
			Versions: versions,
			Error:    fmt.Sprintf("Failed to check for updates: %v", err),
		})
		return
	}

	// If no releases found
	if len(releases) == 0 {
		c.JSON(http.StatusOK, UpdateStatus{
			Versions:        versions,
			LatestVersion:   version.Current,
			UpdateAvailable: false,
			Status:          "Up to Date",
		})
		return
	}

	// Sort releases by version number (newest first)
	sort.Slice(releases, func(i, j int) bool {
		return compareVersions(releases[i].TagName, releases[j].TagName) > 0
	})

	// Get latest release
	latestRelease := releases[0]
	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")
	fmt.Printf("Latest release: %+v\n", latestRelease)
	fmt.Printf("Latest version (after trim): %s\n", latestVersion)

	// Compare versions
	comparison := compareVersions(version.Current, latestVersion)
	fmt.Printf("Version comparison: current=%s, latest=%s, comparison=%d\n", version.Current, latestVersion, comparison)

	var status string
	var updateAvailable bool

	switch {
	case comparison > 0:
		status = "Development Version"
		updateAvailable = false
	case comparison < 0:
		status = "Update Available"
		updateAvailable = true
	default:
		status = "Up to Date"
		updateAvailable = false
	}

	c.JSON(http.StatusOK, UpdateStatus{
		Versions:        versions,
		LatestVersion:   latestVersion,
		UpdateAvailable: updateAvailable,
		Status:          status,
		LatestRelease:   &latestRelease,
	})
}

// fetchGitHubReleases fetches releases from GitHub API
func fetchGitHubReleases(repo string) ([]GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// compareVersions compares two version strings (x.y.z format)
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	// Remove v prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int

		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i]) // Ignore error, treat invalid as 0
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i]) // Ignore error, treat invalid as 0
		}

		if num1 > num2 {
			return 1
		}
		if num1 < num2 {
			return -1
		}
	}

	return 0
}

// UpgradeResponse represents the response for upgrade operation
type UpgradeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// UpgradeRequest represents the upgrade configuration
type UpgradeRequest struct {
	BuildFromSource bool `json:"build_from_source"` // whether to build from source
}

// UpgradeStatus represents the current upgrade status
type UpgradeStatus struct {
	Status    string `json:"status"`    // current status
	Progress  int    `json:"progress"`  // progress percentage
	Error     string `json:"error"`     // error message if any
	Completed bool   `json:"completed"` // whether update is complete
	Log       string `json:"log"`       // update log
}

// GetUpgradeStatus returns the current upgrade status
func GetUpgradeStatus(c *gin.Context) {
	progress, err := services.GetUpdateStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get update status: %v", err),
		})
		return
	}

	// Get update log if available
	log := ""
	if progress.Status != "" {
		if logText, err := services.GetUpdateLog(); err == nil {
			log = logText
		}
	}

	c.JSON(http.StatusOK, UpgradeStatus{
		Status:    progress.Status,
		Progress:  progress.Progress,
		Error:     progress.Error,
		Completed: progress.Completed,
		Log:       log,
	})
}

// PerformUpgrade executes the upgrade process
func PerformUpgrade(c *gin.Context) {
	var req UpgradeRequest
	if err := c.BindJSON(&req); err != nil {
		// Default to not building from source if no request body
		req = UpgradeRequest{
			BuildFromSource: false,
		}
	}

	// Start update via CLI
	if err := services.UpdateSystem(req.BuildFromSource); err != nil {
		c.JSON(http.StatusInternalServerError, UpgradeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UpgradeResponse{
		Success: true,
		Message: "Update started. Check status for progress.",
	})
}
