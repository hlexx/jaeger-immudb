package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ToJsonBytes(input interface{}) []byte {
	marshal, err := json.Marshal(input)
	if err != nil {
		return []byte(``)
	}
	return marshal
}

func FileIsExisted(filename string) bool {
	existed := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		existed = false
	}
	return existed
}
func MakeDir(dir string) error {
	if !FileIsExisted(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			fmt.Println("MakeDir failed:", err)
			return err
		}
	}
	return nil
}

func CopyFile(src, dst string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()
	fi, _ := srcFile.Stat()
	perm := fi.Mode()
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

func CopyDir(srcPath, dstPath string) error {
	if srcInfo, err := os.Stat(srcPath); err != nil {
		return err
	} else {
		if !srcInfo.IsDir() {
			return err
		}
	}
	if desInfo, err := os.Stat(dstPath); err != nil {
		return err
	} else {
		if !desInfo.IsDir() {
			return err
		}
	}
	if strings.TrimSpace(srcPath) == strings.TrimSpace(dstPath) {
		return errors.New("name can't be the same")
	}
	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if path == srcPath {
			return nil
		}
		destNewPath := strings.Replace(path, srcPath, dstPath, -1)
		if !f.IsDir() {
			if strings.Contains(destNewPath, "LOCK") {
				return nil
			}
			_, err = CopyFile(path, destNewPath)
		} else {
			if !FileIsExisted(destNewPath) {
				return MakeDir(destNewPath)
			}
		}

		return nil
	})
	return err
}
