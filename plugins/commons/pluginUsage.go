package commons

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/mitchellh/go-homedir"
)

// InstallPath returns an installation path from the plugin directory.
func (c *Plugin) InstallPath() string {
	return filepath.Join(c.Source, c.Version, fmt.Sprintf("yatas-%s", c.Name))
}

func (c *Plugin) TagName() string {
	if c.Version == "latest" {
		return "latest"
	}
	return fmt.Sprintf("v%s", c.Version)
}

// AssetName returns a name that the asset contained in the release should meet.
// The name must be in a format similar to `yatas-aws_darwin_amd64.zip`.
func (c *Plugin) AssetName() string {
	return fmt.Sprintf("yatas-%s_%s_%s.zip", c.Name, runtime.GOOS, runtime.GOARCH)
}

func (c *Plugin) Validate() error {
	if c.Version != "" && c.Source == "" {
		return fmt.Errorf("plugin `%s`: `source` attribute cannot be omitted when specifying `version`", c.Name)
	}

	if c.Type != "checks" && c.Type != "" && c.Type != "report" && c.Type != "mod" {
		return fmt.Errorf("plugin `%s`: `type` attribute must be either `checks` or `reporting` or `mod`", c.Name)
	}
	if c.Source != "" {
		if c.Version == "" {
			return fmt.Errorf("plugin `%s`: `version` attribute cannot be omitted when specifying `source`", c.Name)
		}

		parts := strings.Split(c.Source, "/")
		// Expected `github.com/owner/repo` format
		if len(parts) != 3 {
			return fmt.Errorf("plugin `%s`: `source` is invalid. Must be in the format `github.com/owner/repo`", c.Name)
		}
		if parts[0] != "github.com" {
			return fmt.Errorf("plugin `%s`: `source` is invalid. Hostname must be `github.com`", c.Name)
		}
		c.SourceOwner = parts[1]
		c.SourceRepo = parts[2]

	}

	return nil
}

func (c *Plugin) Install() (string, error) {

	dir, err := homedir.Expand("~/.yatas.d/plugins")
	if err != nil {
		return "", fmt.Errorf("failed to get plugin dir: %w", err)
	}

	path := filepath.Join(dir, c.InstallPath()+fileExt())
	log.Printf("[DEBUG] Mkdir plugin dir: %s", filepath.Dir(path))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to mkdir to %s: %w", filepath.Dir(path), err)
	}
	fmt.Println("GOGOGOGOG", c.SourceOwner, c.SourceRepo, c.TagName())
	c.Validate()
	fmt.Println("GOGOGOGOG", c.SourceOwner, c.SourceRepo, c.TagName())

	assets, err := c.fetchReleaseAssets()
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitHub releases: %w", err)
	}

	log.Printf("[DEBUG] Download checksums.txt")
	checksumsFile, err := c.downloadToTempFile(assets["checksums.txt"])
	if checksumsFile != nil {
		defer os.Remove(checksumsFile.Name())
	}
	if err != nil {
		return "", fmt.Errorf("failed to download checksums.txt: %s", err)
	}

	log.Printf("[DEBUG] Download %s", c.AssetName())
	zipFile, err := c.downloadToTempFile(assets[c.AssetName()])
	if zipFile != nil {
		defer os.Remove(zipFile.Name())
	}
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %s", c.AssetName(), err)
	}

	if err = extractFileFromZipFile(zipFile, path); err != nil {
		return "", fmt.Errorf("failed to extract binary from %s: %s", c.AssetName(), err)
	}

	log.Printf("[DEBUG] Installed %s successfully", path)
	return path, nil
}

func GetRelease(ctx context.Context, client *github.Client, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
	if tag == "latest" {
		return client.Repositories.GetLatestRelease(ctx, owner, repo)
	} else {
		return client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	}
}

func GetLatestReleaseTag(plugin Plugin) (string, error) {
	ctx := context.Background()
	client := newGitHubClient(ctx)
	plugin.Validate()
	latestRelease, _, err := client.Repositories.GetLatestRelease(ctx, plugin.SourceOwner, plugin.SourceRepo)
	if err != nil {
		return "", err
	}

	return latestRelease.GetName()[1:], nil
}

// fetchReleaseAssets fetches assets from the GitHub release.
// The release is determined by the source path and tag name.
func (c *Plugin) fetchReleaseAssets() (map[string]*github.ReleaseAsset, error) {
	assets := map[string]*github.ReleaseAsset{}

	ctx := context.Background()
	client := newGitHubClient(ctx)

	log.Printf("[DEBUG] Request to https://api.github.com/repos/%s/%s/releases/tags/%s", c.SourceOwner, c.SourceRepo, c.TagName())
	release, _, err := GetRelease(ctx, client, c.SourceOwner, c.SourceRepo, c.TagName())
	if err != nil {
		return assets, err
	}

	for _, asset := range release.Assets {
		log.Printf("[DEBUG] asset found: %s", asset.GetName())
		assets[asset.GetName()] = asset

	}
	return assets, nil
}

// downloadToTempFile download assets from GitHub to a local temp file.
// It is the caller's responsibility to delete the generated the temp file.
func (c *Plugin) downloadToTempFile(asset *github.ReleaseAsset) (*os.File, error) {
	if asset == nil {
		return nil, fmt.Errorf("file not found in the GitHub release. Does the release contain the file with the correct name ?")
	}

	ctx := context.Background()
	client := newGitHubClient(ctx)

	log.Printf("[DEBUG] Request to https://api.github.com/repos/%s/%s/releases/assets/%d", c.SourceOwner, c.SourceRepo, asset.GetID())
	downloader, _, err := client.Repositories.DownloadReleaseAsset(ctx, c.SourceOwner, c.SourceRepo, asset.GetID(), http.DefaultClient)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp("", "yatas-tmp-*")
	if err != nil {
		return nil, err
	}

	fmt.Println(file.Name())
	fmt.Println(downloader)
	if _, err = io.Copy(file, downloader); err != nil {
		return file, err
	}
	downloader.Close()
	if _, err := file.Seek(0, 0); err != nil {
		return file, err
	}

	log.Printf("[DEBUG] Downloaded to %s", file.Name())
	return file, nil
}
