package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// UpdateCheckResponse represents the response for update check
type UpdateCheckResponse struct {
	CurrentVersion  string         `json:"current_version"`
	LatestVersion   string         `json:"latest_version"`
	UpdateAvailable bool           `json:"update_available"`
	LatestRelease   *GitHubRelease `json:"latest_release,omitempty"`
	Error           string         `json:"error,omitempty"`
}

// CheckForUpdates checks for available updates from GitHub releases
func CheckForUpdates(c *gin.Context) {
	// GitHub API URL for releases
	url := "https://api.github.com/repos/griffinwebnet/Vexa/releases"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request to GitHub API
	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, UpdateCheckResponse{
			CurrentVersion: "0.0.1-prealpha",
			Error:          "Failed to check for updates: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Handle 404 (no releases) or other errors
	if resp.StatusCode == 404 {
		c.JSON(http.StatusOK, UpdateCheckResponse{
			CurrentVersion:  "0.0.1-prealpha",
			LatestVersion:   "0.0.1-prealpha",
			UpdateAvailable: false,
			Error:           "No releases found (repository may not have releases yet)",
		})
		return
	}

	if resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, UpdateCheckResponse{
			CurrentVersion: "0.0.1-prealpha",
			Error:          "GitHub API returned status: " + resp.Status,
		})
		return
	}

	// Parse response
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		// For now, return a mock response since we don't have a real repo
		c.JSON(http.StatusOK, UpdateCheckResponse{
			CurrentVersion:  "0.0.1-prealpha",
			LatestVersion:   "0.0.1-prealpha",
			UpdateAvailable: false,
			Error:           "No releases available (development mode)",
		})
		return
	}

	// If no releases found, return current version
	if len(releases) == 0 {
		c.JSON(http.StatusOK, UpdateCheckResponse{
			CurrentVersion:  "0.0.1-prealpha",
			LatestVersion:   "0.0.1-prealpha",
			UpdateAvailable: false,
		})
		return
	}

	// Filter out pre-releases and sort by version
	var stableReleases []GitHubRelease
	for _, release := range releases {
		if !strings.Contains(release.TagName, "pre") && !strings.Contains(release.TagName, "alpha") && !strings.Contains(release.TagName, "beta") {
			stableReleases = append(stableReleases, release)
		}
	}

	// Sort releases by version (newest first)
	sort.Slice(stableReleases, func(i, j int) bool {
		return compareVersions(stableReleases[i].TagName, stableReleases[j].TagName) > 0
	})

	currentVersion := "0.0.1-prealpha"
	updateAvailable := false
	var latestRelease *GitHubRelease

	if len(stableReleases) > 0 {
		latestRelease = &stableReleases[0]
		latestVersion := latestRelease.TagName

		// Compare versions (skip if current is pre-release)
		if !strings.Contains(currentVersion, "pre") && !strings.Contains(currentVersion, "alpha") && !strings.Contains(currentVersion, "beta") {
			updateAvailable = compareVersions(latestVersion, currentVersion) > 0
		}
	}

	response := UpdateCheckResponse{
		CurrentVersion:  currentVersion,
		UpdateAvailable: updateAvailable,
		LatestRelease:   latestRelease,
	}

	if latestRelease != nil {
		response.LatestVersion = latestRelease.TagName
	}

	c.JSON(http.StatusOK, response)
}

// compareVersions compares two semantic version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split versions into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Ensure both have at least 3 parts (major.minor.patch)
	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}

	// Compare each part
	for i := 0; i < 3; i++ {
		num1, err1 := strconv.Atoi(parts1[i])
		num2, err2 := strconv.Atoi(parts2[i])

		// If parsing fails, do string comparison
		if err1 != nil || err2 != nil {
			if parts1[i] > parts2[i] {
				return 1
			} else if parts1[i] < parts2[i] {
				return -1
			}
			continue
		}

		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}

	return 0
}
