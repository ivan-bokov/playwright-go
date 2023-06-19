package playwright_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/h2non/filetype"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestVideoShouldWork(t *testing.T) {
	recordVideoDir := t.TempDir()
	BeforeEach(t, playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir: playwright.String(recordVideoDir),
			Size: &playwright.RecordVideoSize{
				Width:  playwright.Int(500),
				Height: playwright.Int(400),
			},
		},
	})
	defer AfterEach(t)
	_, err := page.Goto(server.PREFIX + "/grid.html")
	require.NoError(t, err)
	_, err = page.Reload()
	require.NoError(t, err)
	_, err = page.Reload()
	require.NoError(t, err)
	page.WaitForTimeout(1000) // make sure video has some data
	require.NoError(t, context.Close())

	path, err := page.Video().Path()
	require.NoError(t, err)
	files, err := os.ReadDir(recordVideoDir)
	require.NoError(t, err)
	require.Equal(t, len(files), 1)
	videoFileLocation := filepath.Join(recordVideoDir, files[0].Name())
	require.Equal(t, videoFileLocation, path)
	require.FileExists(t, videoFileLocation)
	content, err := os.ReadFile(videoFileLocation)
	require.NoError(t, err)
	require.True(t, filetype.IsVideo(content))
	tmpFile := filepath.Join(t.TempDir(), "test.webm")
	require.NoError(t, page.Video().SaveAs(tmpFile))
	require.FileExists(t, tmpFile)
	require.NoError(t, page.Video().Delete())
	require.NoFileExists(t, videoFileLocation)
}

func TestVideo(t *testing.T) {
	t.Run("should expose video path", func(t *testing.T) {
		recordVideoDir := t.TempDir()
		BeforeEach(t, playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(recordVideoDir),
				Size: &playwright.RecordVideoSize{
					Width:  playwright.Int(500),
					Height: playwright.Int(400),
				},
			},
		})
		defer AfterEach(t)
		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, recordVideoDir)
		page.WaitForTimeout(1000)
		require.NoError(t, page.Context().Close())
	})

	t.Run("should work when access video after close page", func(t *testing.T) {
		recordVideoDir := t.TempDir()
		BeforeEach(t, playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(recordVideoDir),
				Size: &playwright.RecordVideoSize{
					Width:  playwright.Int(500),
					Height: playwright.Int(400),
				},
			},
		})
		defer AfterEach(t)
		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		page.WaitForTimeout(1000)
		require.NoError(t, page.Close())
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, recordVideoDir)
		require.FileExists(t, path)
	})

	t.Run("video should not exist when delete before close page", func(t *testing.T) {
		recordVideoDir := t.TempDir()
		BeforeEach(t, playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(recordVideoDir),
				Size: &playwright.RecordVideoSize{
					Width:  playwright.Int(500),
					Height: playwright.Int(400),
				},
			},
		})
		defer AfterEach(t)
		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		page.WaitForTimeout(1000)
		require.NoError(t, page.Close())
		require.NoError(t, video.Delete())
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, recordVideoDir)
		require.NoFileExists(t, path)
	})

	t.Run("video should not exist when no dir specified", func(t *testing.T) {
		BeforeEach(t)
		defer AfterEach(t)
		_, err := page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.Error(t, err)
		require.Empty(t, path)
		tmpFile := filepath.Join(t.TempDir(), "test.webm")
		require.Error(t, video.SaveAs(tmpFile))
		require.NoError(t, page.Context().Close())
		require.Error(t, video.SaveAs(tmpFile))
		require.NoError(t, video.Delete())
	})

	t.Run("record video to path persistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		BeforeEach(t)
		defer AfterEach(t)
		require.NoError(t, context.Close())

		bt := browser.BrowserType()

		context, err := bt.LaunchPersistentContext(tmpDir, playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless: playwright.Bool(os.Getenv("HEADFUL") == ""),
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(tmpDir),
			},
		})
		require.NoError(t, err)
		page := context.Pages()[0]
		_, err = page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		video := page.Video()
		require.NotNil(t, video)
		path, err := video.Path()
		require.NoError(t, err)
		require.Contains(t, path, tmpDir)
		page.WaitForTimeout(1000)
		require.NoError(t, context.Close())
		require.FileExists(t, path)
	})

	t.Run("remote server should work with saveas", func(t *testing.T) {
		tmpDir := t.TempDir()
		BeforeEach(t)
		defer AfterEach(t)
		remoteServer, err := newRemoteServer()
		require.NoError(t, err)
		defer remoteServer.Close()
		browser, err := browserType.Connect(remoteServer.url)
		require.NoError(t, err)
		require.NotNil(t, browser)
		browser_context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			RecordVideo: &playwright.RecordVideo{
				Dir: playwright.String(tmpDir),
			},
		})
		require.NoError(t, err)
		page, err = browser_context.NewPage()
		require.NoError(t, err)
		_, err = page.Goto(server.PREFIX + "/grid.html")
		require.NoError(t, err)
		page.WaitForTimeout(1000)
		video := page.Video()
		_, err = video.Path()
		require.ErrorContains(t, err, "Path is not available when connecting remotely")
		tmpFile := filepath.Join(t.TempDir(), "test.webm")
		require.ErrorContains(t, video.SaveAs(tmpFile), "Page is not yet closed.")
		require.NoError(t, browser_context.Close())
		require.NoError(t, video.SaveAs(tmpFile))
		require.FileExists(t, tmpFile)
	})
}
