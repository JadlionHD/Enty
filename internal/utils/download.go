package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (u *utils) DownloadFile(name string, filename string, url string, buf int32) (err error) {

	dirPath := filepath.Join(u.GetPwd(), PATH_TEMP)
	exist := u.IsDirExist(dirPath)

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

		// Close file first, or it will throw "it is being used by another process."
		out.Close()
		filePath := filepath.Join(u.GetPwd(), PATH_TEMP, filename)

		remErr := os.Remove(filePath)
		log.Printf("File cancelled: %s with err: %s", filePath, remErr)
	})
	defer runtime.EventsOffAll(u.ctx)

	for {
		select {
		case <-ctx.Done():
			runtime.EventsEmit(u.ctx, "download-cancelled", name)
			runtime.EventsOffAll(u.ctx)
			filePath := filepath.Join(u.GetPwd(), PATH_TEMP, filename)
			log.Printf("Cleanup partial file: %s", filePath)
			os.Remove(filePath) // Clean up partial file
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
