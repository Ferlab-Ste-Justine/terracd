package fs

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

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

func AssurePrivateDir(dir string) {
	mkdirErr := os.MkdirAll(dir, 0770)
	if mkdirErr != nil {
		panic(mkdirErr.Error())
	}
}

func CopyPrivateFile(src string, dest string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, input, 0660)
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

func FindFiles(dir string, pattern string) ([]string, error) {
	matches := []string{}

	elems, readDirErr := ioutil.ReadDir(dir)
	if readDirErr != nil {
		return matches, readDirErr
	}

	for _, elem := range elems {
		elemPath := path.Join(dir, elem.Name())
		elemInfo, err := os.Stat(elemPath)
		if err != nil {
			return matches, err
		}

		if elemInfo.IsDir() {
			subMatches, err := FindFiles(elemPath, pattern)
			if err != nil {
				return matches, err
			}
			matches = append(matches, subMatches...)
		} else {
			match, err := filepath.Match(path.Join(dir, pattern), elemPath)
			if err != nil {
				return matches, err
			}
			if match {
				matches = append(matches, elemPath)
			}
		}
	}

	return matches, nil
}