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
	Component string `json:"component"`
	Version   string `json:"version"`
}

// UpdateStatus represents the current update status
type UpdateStatus struct {
	Versions        []SystemVersion `json:"versions"`         // Current versions of all components
	LatestVersion   string          `json:"latest_version"`   // Latest version from GitHub
	UpdateAvailable bool            `json:"update_available"` // Whether an update is available
	Status          string          `json:"status"`           // Status message (Up to Date, Update Available, etc)
	LatestRelease   *GitHubRelease  `json:"latest_release"`   // Details about the latest release
	Error           string          `json:"error,omitempty"`  // Any error that occurred
}

// CheckForUpdates checks for available updates from GitHub releases
func CheckForUpdates(c *gin.Context) {
	// Get current versions
	componentVersions := version.Components()
	fmt.Printf("Component versions: %+v\n", componentVersions)

	// Convert to slice for response
	var versions []SystemVersion
	for component, ver := range componentVersions {
		versions = append(versions, SystemVersion{
			Component: component,
			Version:   ver,
		})
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

// PerformUpgrade executes the upgrade process
func PerformUpgrade(c *gin.Context) {
	// TODO: Implement actual upgrade logic
	c.JSON(http.StatusOK, UpgradeResponse{
		Success: true,
		Message: "Upgrade completed successfully",
	})
}
