package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (u *utils) DownloadFile(name string, filename string, url string, buf int32) (err error) {
	exist := u.IsDirExist(PATH_TEMP)

	ctx, cancel := context.WithCancel(u.ctx)
	defer cancel()

	if !exist {
		u.Mkdir(PATH_TEMP)
	}

	// Create the file
	out, err := os.Create(fmt.Sprintf("%s/%s", PATH_TEMP, filename))
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	totalBytes := resp.ContentLength
	var downloadedBytes int64 = 0

	// Create a buffer to write with 32 KB chunks
	buffer := make([]byte, buf*1024)

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	runtime.EventsEmit(u.ctx, "start-download-file", name, totalBytes)

	runtime.EventsOnce(u.ctx, "cancel-download-file", func(optionalData ...interface{}) {
		// Cancel all downloads if no filename provided
		// Or cancel specific download if filename matches
		if len(optionalData) == 0 {
			cancel()
			runtime.EventsEmit(u.ctx, "download-cancelled", name)
		} else if filename, ok := optionalData[0].(string); ok && filename == name {
			cancel()
			runtime.EventsEmit(u.ctx, "download-cancelled", name)
		}
	})
	defer runtime.EventsOffAll(u.ctx)

	for {
		select {
		case <-ctx.Done():
			runtime.EventsEmit(u.ctx, "download-cancelled", name)
			os.Remove(fmt.Sprintf("%s/%s", PATH_TEMP, name)) // Clean up partial file
			return ctx.Err()
		default:
			n, err := io.ReadFull(resp.Body, buffer)

			if n > 0 {
				_, writeErr := out.Write(buffer[:n])
				if writeErr != nil {
					return fmt.Errorf("write error: %v", writeErr)
				}
				downloadedBytes += int64(n)

				runtime.EventsEmit(u.ctx, "download-file", name, totalBytes, downloadedBytes)
			}

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				runtime.EventsEmit(u.ctx, "finish-download-file", name)
				return nil
			}
			if err != nil {
				return fmt.Errorf("read error: %v", err)
			}
		}
	}

}
