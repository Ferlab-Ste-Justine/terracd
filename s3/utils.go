package s3

import (
	"context"

	minio "github.com/minio/minio-go/v7"
)

func KeyExists(bucket string, key string, client *minio.Client) (bool, error) {
	_, err := client.StatObject(context.Background(), bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}