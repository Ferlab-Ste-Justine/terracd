package s3

import(
	"context"
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
		return delErr
	}
	
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
			return obj.Err
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
				resultsCh <- readErr
				return
			}

			fWr, fWrErr := os.Create(path.Join(fsPath, destRelPath))
			if fWrErr != nil {
				resultsCh <- fWrErr
				return
			}
			defer fWr.Close()

			_, copyErr := io.Copy(fWr, objRead)
			if copyErr != nil {
				resultsCh <- copyErr
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
			return obj.Err
		}
	
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			rmErr := conn.RemoveObject(context.Background(), s3Conf.Bucket, obj.Key, minio.RemoveObjectOptions{})
			if rmErr != nil {
				resultsCh <- rmErr
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
		return connErr
	}

	files, filesErr := fs.FindFiles(fsPath, "*")
	if filesErr != nil {
		return filesErr
	}

	resultsCh := make(chan error)
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			fInfo, fInfoErr := os.Lstat(file)
			if fInfoErr != nil {
				resultsCh <- fInfoErr
				return
			}

			fHandle, openErr := os.Open(file)
			if openErr != nil {
				resultsCh <- openErr
				return
			}
			defer fHandle.Close()

			destRelPath, relPathErr := filepath.Rel(fsPath, file)
			if relPathErr != nil {
				resultsCh <- relPathErr
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
			resultsCh <- putErr
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