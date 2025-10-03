package version

// Version information for the entire Vexa system
const (
	// Current version of the entire system (API and Web)
	Current = "0.1.32"
)

// Components returns version information for all system components
func Components() map[string]string {
	return map[string]string{
		"api": Current,
		"web": Current,
	}
}
