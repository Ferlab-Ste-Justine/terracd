package fs

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Paths struct {
	Root    string
	Repos   string
	Backend string
	State   string
	FsStore string
	Cache   string
	Work    string
}

func GetPaths(rootDir string) Paths {
	return Paths{
		Root: rootDir,
		Repos: path.Join(rootDir, "repos"),
		Backend: path.Join(rootDir, "backend"),
		State: path.Join(rootDir, "state"),
		FsStore: path.Join(rootDir, "fs-store"),
		Cache: path.Join(rootDir, "cache"),
		Work: path.Join(rootDir, "work"),
	}
}

func GetFileSha256(src string) (string, error) {
	fHandle, handleErr := os.Open(src)
	if handleErr != nil {
	  return "", handleErr
	}
	defer fHandle.Close()
  
	hash := sha256.New()
	_, copyErr := io.Copy(hash, fHandle)
	if copyErr != nil {
		return "", copyErr
	}
  
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func EnsureContainingDirExists(path string) error {
	dir := filepath.Dir(path)
	
	if dir != "." && dir != "/" {
		return AssurePrivateDir(dir)
	}
	
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return true, err
		}

		return false, nil
	}

	return true, nil
}

func AssurePrivateDir(dir string) error {
	return os.MkdirAll(dir, 0770)
}

func CopyPrivateFile(src string, dest string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, input, 0770)
	if err != nil {
		return err
	}

	return nil
}

func CopyDir(destDir string, sourceDir string) error {
	elems, readDirErr := ioutil.ReadDir(sourceDir)
	if readDirErr != nil {
		return readDirErr
	}

	for _, elem := range elems {
		src := path.Join(sourceDir, elem.Name())
		dest := path.Join(destDir, elem.Name())

		srcInfo, err := os.Stat(src)
		if err != nil {
			return err
		}

		if srcInfo.IsDir() {
			err := os.MkdirAll(dest, 0770)
			if err != nil {
				return err
			}

			err = CopyDir(dest, src)
			if err != nil {
				return err
			}
		} else {
			err := CopyPrivateFile(src, dest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func MergeDirs(destDir string, sourceDirs []string) error {
	for _, sourceDir := range sourceDirs {
		err := CopyDir(destDir, sourceDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func FindFiles(src string, pattern string) ([]string, error) {
	matches := []string{}

	err := filepath.Walk(src, func(fPath string, fInfo fs.FileInfo, fErr error) error {
		if fErr != nil {
			return fErr
		}

		fMode := fInfo.Mode()
		if fMode & fs.ModeSymlink != 0 {
			link, linkErr := os.Readlink(fPath)
			if linkErr != nil {
				return linkErr
			}

			if !path.IsAbs(link) {
				link = path.Join(path.Dir(fPath), link)
			}

			stat, statErr := os.Stat(link)
			if statErr != nil {
				if strings.Contains(statErr.Error(),  "too many levels of symbolic links") {
					return nil
				}

				return statErr
			}

			if stat.IsDir() {
				return nil
			}
		}

		if fInfo.IsDir() {
			return nil
		}

		match, mErr := filepath.Match(path.Join(filepath.Dir(fPath), pattern), fPath)
		if mErr != nil {
			return mErr
		}

		if match {
			matches = append(matches, fPath)
		}

		return nil
	})

	return matches, err
}