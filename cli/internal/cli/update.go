package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Self-update oraculo binary from GitHub Releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate()
		},
	}
}

const releasesAPI = "https://api.github.com/repos/lucas-stellet/oraculo/releases/latest"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpdate() error {
	fmt.Println("[oraculo] Checking for updates...")

	resp, err := http.Get(releasesAPI)
	if err != nil {
		return fmt.Errorf("checking releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("parsing release: %w", err)
	}

	if release.TagName == "" {
		fmt.Println("[oraculo] No releases found.")
		return nil
	}

	fmt.Printf("[oraculo] Latest release: %s\n", release.TagName)

	// Find matching asset
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	wantSuffix := fmt.Sprintf("_%s_%s.tar.gz", goos, goarch)

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, wantSuffix) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s in release %s", goos, goarch, release.TagName)
	}

	fmt.Printf("[oraculo] Downloading %s/%s binary...\n", goos, goarch)

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving self path: %w", err)
	}

	dlResp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != 200 {
		return fmt.Errorf("download returned %d", dlResp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "oraculo-update-*.tar.gz")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, dlResp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("writing download: %w", err)
	}
	tmp.Close()

	// Extract binary from tarball and replace self
	if err := extractAndReplace(tmp.Name(), self); err != nil {
		return fmt.Errorf("extracting update: %w", err)
	}

	fmt.Printf("[oraculo] Updated: %s\n", self)
	return nil
}

func extractAndReplace(tarPath, destPath string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("oraculo binary not found in tarball")
		}
		if err != nil {
			return err
		}
		if hdr.Name == "oraculo" {
			tmp, err := os.CreateTemp("", "oraculo-bin-*")
			if err != nil {
				return err
			}
			if _, err := io.Copy(tmp, tr); err != nil {
				tmp.Close()
				os.Remove(tmp.Name())
				return err
			}
			tmp.Close()
			if err := os.Chmod(tmp.Name(), 0755); err != nil {
				os.Remove(tmp.Name())
				return err
			}
			if err := os.Rename(tmp.Name(), destPath); err != nil {
				os.Remove(tmp.Name())
				return err
			}
			return nil
		}
	}
}
