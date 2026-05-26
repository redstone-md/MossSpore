package version

var (
	// Version is the current version of MossSpore, set at build time.
	Version = "0.1.0-dev"

	// Commit is the git commit hash, set at build time via -ldflags.
	Commit = "unknown"

	// Date is the build timestamp, set at build time via -ldflags.
	Date = "unknown"
)

// Info returns a human-readable version string.
func Info() string {
	s := "MossSpore " + Version
	if Commit != "unknown" {
		s += " (" + Commit[:min(8, len(Commit))] + ")"
	}
	if Date != "unknown" {
		s += " built " + Date
	}
	return s
}
