package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	githubAPI    = "https://api.github.com/repos/%s/%s/releases/latest"
	sentinelFile = "update.json"
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// state is the sentinel file content for crash recovery.
type state struct {
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	BackupPath  string `json:"backup_path"`
	Timestamp   string `json:"timestamp"`
}

// Updater handles automatic binary updates.
type Updater struct {
	repoOwner  string
	repoName   string
	currentVer string
	dataDir    string
	httpClient *http.Client
	mu         sync.Mutex
}

// NewUpdater creates an updater. currentVersion is the running version
// (e.g. "0.1.1"). dataDir is where update state is stored.
func NewUpdater(currentVersion, dataDir string) *Updater {
	_ = os.MkdirAll(dataDir, 0700)
	return &Updater{
		repoOwner:  "redstone-md",
		repoName:   "MossSpore",
		currentVer: strings.TrimSuffix(currentVersion, "-dev"),
		dataDir:    dataDir,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Check fetches the latest release from GitHub.
func (u *Updater) Check(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf(githubAPI, u.repoOwner, u.repoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// IsNewer returns true if the release is a newer version than current.
func (u *Updater) IsNewer(release *Release) bool {
	tag := strings.TrimPrefix(release.TagName, "v")
	return tag != "" && semverGreater(tag, u.currentVer)
}

// FetchAndVerify downloads the binary for the current platform and verifies
// its SHA256 checksum. Returns the path to the downloaded temp file.
func (u *Updater) FetchAndVerify(ctx context.Context, release *Release) (string, error) {
	assetName := fmt.Sprintf("mossspore-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var binAsset, sumAsset *Asset
	for i := range release.Assets {
		a := &release.Assets[i]
		if a.Name == assetName {
			binAsset = a
		}
		if a.Name == "checksums.txt" {
			sumAsset = a
		}
	}
	if binAsset == nil {
		return "", fmt.Errorf("no binary for %s/%s in release", runtime.GOOS, runtime.GOARCH)
	}

	var expectedHash string
	if sumAsset != nil {
		expectedHash, _ = u.fetchChecksum(ctx, sumAsset, assetName)
	}

	tmp, err := os.CreateTemp("", "mossspore-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, binAsset.BrowserDownloadURL, nil)
	if err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", err
	}
	resp, err := u.httpClient.Do(req)
	if err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", err
	}
	defer resp.Body.Close()

	hash := sha256.New()
	w := io.MultiWriter(tmp, hash)

	if _, err := io.Copy(w, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("download: %w", err)
	}
	tmp.Close()

	got := hex.EncodeToString(hash.Sum(nil))
	if expectedHash != "" && got != expectedHash {
		os.Remove(tmpPath)
		return "", fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, got)
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(tmpPath, 0755)
	}

	return tmpPath, nil
}

// Apply replaces the current binary with the downloaded one and writes a
// sentinel for crash recovery. targetVersion is the expected version after
// the update (e.g. "0.1.2"). The file at tmpPath is consumed (removed).
func (u *Updater) Apply(tmpPath, targetVersion string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	backupPath := currentPath + ".bak"

	if err := copyFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	if err := copyFile(tmpPath, currentPath); err != nil {
		_ = os.Remove(backupPath)
		return fmt.Errorf("replace: %w", err)
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(currentPath, 0755)
	}

	s := state{
		FromVersion: u.currentVer,
		ToVersion:   targetVersion,
		BackupPath:  backupPath,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := u.writeSentinel(s); err != nil {
		_ = copyFile(backupPath, currentPath)
		_ = os.Remove(backupPath)
		return fmt.Errorf("sentinel: %w", err)
	}

	_ = os.Remove(tmpPath)
	return nil
}

// Restart re-execs into the current binary. On Unix this uses syscall.Exec
// so the process image is replaced in-place. On Windows it exits cleanly.
func (u *Updater) Restart() {
	currentPath, err := os.Executable()
	if err != nil {
		log.Printf("[update] restart: %v", err)
		return
	}

	if runtime.GOOS != "windows" {
		log.Printf("[update] re-exec into %s", currentPath)
		if err := syscall.Exec(currentPath, os.Args, os.Environ()); err != nil {
			log.Printf("[update] re-exec failed: %v", err)
		}
	}

	// Windows fallback: exit and let service manager restart us
	log.Printf("[update] exiting for service-manager restart")
	os.Exit(0)
}

// DoFullUpdate runs the complete update cycle: check → fetch → apply → restart.
// Returns nil if no update is needed or if the update was applied (process will
// exit or re-exec). Returns an error if something went wrong.
func (u *Updater) DoFullUpdate(ctx context.Context) error {
	release, err := u.Check(ctx)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	if !u.IsNewer(release) {
		return nil
	}

	log.Printf("[update] found %s (have %s)", release.TagName, u.currentVer)

	tmpPath, err := u.FetchAndVerify(ctx, release)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	target := strings.TrimPrefix(release.TagName, "v")
	if err := u.Apply(tmpPath, target); err != nil {
		return fmt.Errorf("apply: %w", err)
	}

	log.Printf("[update] updated to %s, restarting", release.TagName)
	u.Restart()
	return nil
}

// CleanupOnStart should be called at the very beginning of startup. It checks
// for a pending-update sentinel and either confirms the update or rolls back.
func (u *Updater) CleanupOnStart() error {
	s, err := u.readSentinel()
	if err != nil {
		return nil
	}

	currentPath, err := os.Executable()
	if err != nil {
		return err
	}

	if u.currentVer == s.ToVersion {
		_ = os.Remove(s.BackupPath)
		u.removeSentinel()
		log.Printf("[update] update to %s confirmed, cleaned up", s.ToVersion)
		return nil
	}

	log.Printf("[update] version mismatch: have %s, expected %s — rolling back",
		u.currentVer, s.ToVersion)

	if err := copyFile(s.BackupPath, currentPath); err != nil {
		return fmt.Errorf("rollback copy: %w", err)
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(currentPath, 0755)
	}

	u.removeSentinel()
	_ = os.Remove(s.BackupPath)

	log.Printf("[update] rollback complete, re-exec into old binary")
	if runtime.GOOS != "windows" {
		if err := syscall.Exec(currentPath, os.Args, os.Environ()); err != nil {
			return fmt.Errorf("rollback re-exec: %w", err)
		}
	}

	os.Exit(1)
	return nil
}

// --- helpers ---

func (u *Updater) sentinelPath() string {
	return filepath.Join(u.dataDir, sentinelFile)
}

func (u *Updater) writeSentinel(s state) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(u.sentinelPath(), raw, 0600)
}

func (u *Updater) readSentinel() (*state, error) {
	raw, err := os.ReadFile(u.sentinelPath())
	if err != nil {
		return nil, err
	}
	var s state
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (u *Updater) removeSentinel() {
	_ = os.Remove(u.sentinelPath())
}

func (u *Updater) stripV(s string) string {
	return strings.TrimPrefix(s, "v")
}

func (u *Updater) fetchChecksum(ctx context.Context, asset *Asset, targetName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == targetName {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("%s not found in checksums", targetName)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// semverGreater returns true if a > b using simple numeric comparison.
func semverGreater(a, b string) bool {
	pa, pb := splitParts(a), splitParts(b)
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] > pb[i] {
			return true
		}
		if pa[i] < pb[i] {
			return false
		}
	}
	return len(pa) > len(pb)
}

func splitParts(v string) []int {
	var out []int
	for _, s := range strings.Split(v, ".") {
		s = strings.TrimLeft(s, "0")
		if s == "" {
			out = append(out, 0)
			continue
		}
		// Strip non-numeric suffix like "beta", "rc1"
		for i, r := range s {
			if r < '0' || r > '9' {
				s = s[:i]
				break
			}
		}
		if n, err := strconv.Atoi(s); err == nil {
			out = append(out, n)
		} else {
			out = append(out, 0)
		}
	}
	return out
}
