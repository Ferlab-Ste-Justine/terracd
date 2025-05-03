package s3

import(
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	
	minio "github.com/minio/minio-go/v7"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
)

func SyncToFs(s3Conf S3ClientConfig, fsPath string) error {
	delErr := fs.EnsureDirectoryNotExits(fsPath)
	if delErr != nil {
		return errors.New(fmt.Sprintf("Error cleaning up target s3 sync directory: %s", delErr.Error()))
	}
	
	conn, connErr := Connect(s3Conf)
	if connErr != nil {
		return errors.New(fmt.Sprintf("Error connecting to s3 store: %s", connErr.Error()))
	}

	objCh := conn.ListObjects(context.Background(), s3Conf.Bucket, minio.ListObjectsOptions{
		Prefix: s3Conf.Path,
		Recursive: true,
	})

	resultsCh := make(chan error)
	var wg sync.WaitGroup

	for obj := range objCh {
		if obj.Err != nil {
			return errors.New(fmt.Sprintf("Error iterating over bucket '%s' objects: %s", s3Conf.Bucket, obj.Err.Error()))
		}
		
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			prefix := s3Conf.Path
			if prefix != "" {
				if !strings.HasSuffix(prefix, "/") {
					prefix = prefix + "/"
				}
			}

			destRelPath := strings.TrimLeft(key, prefix)

			objRead, readErr := conn.GetObject(context.Background(), s3Conf.Bucket, key, minio.GetObjectOptions{})
			if readErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error reading object '%s': %s", key, readErr.Error()))
				return
			}

			contDirErr := fs.EnsureContainingDirExists(path.Join(fsPath, destRelPath))
			if contDirErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error creating directory for destination path '%s' to copy object: %s", destRelPath, contDirErr.Error()))
				return
			}

			fWr, fWrErr := os.Create(path.Join(fsPath, destRelPath))
			if fWrErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error creating destination path '%s' to copy object: %s", destRelPath, fWrErr.Error()))
				return
			}
			defer fWr.Close()

			_, copyErr := io.Copy(fWr, objRead)
			if copyErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error copying s3 key '%s' into fs path '%s': %s", key, destRelPath, copyErr.Error()))
				return
			}

			resultsCh <- nil
		}(obj.Key)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var err error
	for resErr := range resultsCh {
		if resErr != nil {
			err = resErr
		}
	}

	return err
}

func clearS3Path(s3Conf S3ClientConfig) error {
	conn, connErr := Connect(s3Conf)
	if connErr != nil {
		return connErr
	}

	objCh := conn.ListObjects(context.Background(), s3Conf.Bucket, minio.ListObjectsOptions{
		Prefix: s3Conf.Path,
		Recursive: true,
	})

	resultsCh := make(chan error)
	var wg sync.WaitGroup

	for obj := range objCh {
		if obj.Err != nil {
			return errors.New(fmt.Sprintf("Error iterating over bucket '%s' objects: %s", s3Conf.Bucket, obj.Err.Error()))
		}
	
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			rmErr := conn.RemoveObject(context.Background(), s3Conf.Bucket, obj.Key, minio.RemoveObjectOptions{})
			if rmErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error deleting s3 object '%s': %s", key, rmErr.Error()))
				return
			}
		
			resultsCh <- nil
		}(obj.Key)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	
	var err error
	for resErr := range resultsCh {
		if resErr != nil {
			err = resErr
		}
	}

	return err
}

func SyncFromFs(s3Conf S3ClientConfig, fsPath string) error {
	clearErr := clearS3Path(s3Conf)
	if clearErr != nil {
		return clearErr
	}

	conn, connErr := Connect(s3Conf)
	if connErr != nil {
		return errors.New(fmt.Sprintf("Error connecting to s3 store: %s", connErr.Error()))
	}

	files, filesErr := fs.FindFiles(fsPath, "*")
	if filesErr != nil {
		return errors.New(fmt.Sprintf("Error listing files in path '%s': %s", fsPath, filesErr.Error()))
	}

	resultsCh := make(chan error)
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			fInfo, fInfoErr := os.Lstat(file)
			if fInfoErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error gettings stats on source file '%s': %s", file, fInfoErr.Error()))
				return
			}

			fHandle, openErr := os.Open(file)
			if openErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error opening source file '%s': %s", file, openErr.Error()))
				return
			}
			defer fHandle.Close()

			destRelPath, relPathErr := filepath.Rel(fsPath, file)
			if relPathErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error getting relative path portion of source file '%s': %s", file, relPathErr.Error()))
				return
			}
			
			_, putErr := conn.PutObject(
				context.Background(),
				s3Conf.Bucket,
				path.Join(s3Conf.Path, destRelPath),
				fHandle,
				fInfo.Size(),
				minio.PutObjectOptions{},
			)
			if putErr != nil {
				resultsCh <- errors.New(fmt.Sprintf("Error copying source file '%s' to s3: %s", file, putErr.Error()))
				return
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()
	
	var err error
	for resErr := range resultsCh {
		if resErr != nil {
			err = resErr
		}
	}

	return err
}