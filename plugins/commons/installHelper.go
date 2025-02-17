package commons

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

func extractFileFromZipFile(zipFile *os.File, savePath string) error {
	zipFileStat, err := zipFile.Stat()
	if err != nil {
		return err
	}
	zipReader, err := zip.NewReader(zipFile, zipFileStat.Size())
	if err != nil {
		return err
	}

	var reader io.ReadCloser
	for _, f := range zipReader.File {
		log.Printf("[DEBUG] file found in zip: %s", f.Name)
		if f.Name != filepath.Base(savePath) {
			continue
		}

		reader, err = f.Open()
		if err != nil {
			return err
		}
		break
	}
	if reader == nil {
		return fmt.Errorf("file not found. Does the zip contain %s ?", filepath.Base(savePath))
	}

	file, err := os.OpenFile(savePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(file.Name())
		return err
	}

	return nil
}

func newGitHubClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}

	// log.Printf("[DEBUG] GITHUB_TOKEN set, plugin requests to the GitHub API will be authenticated")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return github.NewClient(oauth2.NewClient(ctx, ts))
}

func fileExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
