package download

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/GoCEP/internal/progressBar"
	"github.com/mitchellh/colorstring"
)

func File(url string, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("/tmp", "go-cep-download-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if cerr := tmpFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			os.Remove(tmpFile.Name())
		}
	}()

	text := fmt.Sprintf("[cyan][1/3][reset] Downloading cep data to %s", tmpFile.Name())

	bar := progressBar.Create(int(resp.ContentLength), text)

	_, err = io.Copy(io.MultiWriter(tmpFile, bar), resp.Body)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); err == nil {
		oldPath := filePath + ".old"

		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Remove(oldPath); err != nil {
				return err
			}
			colorizedString := colorstring.Color("[cyan][2/3][reset] Deleted existing .old file " + oldPath)
			fmt.Println(colorizedString)
		}

		if err := os.Rename(filePath, oldPath); err != nil {
			return err
		}
		colorizedString := colorstring.Color("[cyan][2/3][reset] Moved existing file " + filePath + " to " + oldPath)
		fmt.Println(colorizedString)
	}

  src, err := os.Open(tmpFile.Name())
  if err != nil {
    return err
  }
  defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	fileInfo, err := tmpFile.Stat()
	if err != nil {
		return err
	}

	size := fileInfo.Size()

	text = fmt.Sprintf("[cyan][2/3][reset] Saving %s", filePath)
	bar = progressBar.Create(int(size), text)

	if _, err := io.Copy(io.MultiWriter(dst, bar), src); err != nil {
		return err
	}

	return nil
}
