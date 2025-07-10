package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (u *utils) DownloadFile(name string, filepath string, url string) (err error) {

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		//not exist

		os.Mkdir("temp", os.ModePerm)

	}

	// Create the file
	out, err := os.Create(fmt.Sprintf("temp/%s", name))
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	totalBytes := resp.ContentLength
	var downloadedBytes int64 = 0

	// Create a buffer to write with 32 KB chunks
	buffer := make([]byte, 32*1024)

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	runtime.EventsEmit(u.ctx, "start-download-file", name)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			downloadedBytes += int64(n)
			runtime.EventsEmit(u.ctx, "download-file", name, totalBytes, downloadedBytes)
			// fmt.Printf("\rDownloading... %d%% complete", 100*downloadedBytes/totalBytes)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	runtime.EventsEmit(u.ctx, "finish-download-file", name)

	return nil
}
