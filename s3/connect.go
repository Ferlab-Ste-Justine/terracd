package s3

import (
	"crypto/tls"
    "crypto/x509"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func getTlsConfigs(s3Conf S3ClientConfig) (*tls.Config, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: false,
	}

	//CA cert
	if s3Conf.Auth.CaCert != "" {
		caCertContent, err := ioutil.ReadFile(s3Conf.Auth.CaCert)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to read root certificate file: %s", err.Error()))
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCertContent)
		if !ok {
			return nil, errors.New("Failed to parse root certificate authority")
		}
		(*tlsConf).RootCAs = roots
	}

	return tlsConf, nil
}

func Connect(s3Conf S3ClientConfig) (*minio.Client, error) {
	tlsConf, tlsConfErr := getTlsConfigs(s3Conf)
	if tlsConfErr != nil {
		return nil, tlsConfErr
	}

	return minio.New(s3Conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Conf.Auth.AccessKey, s3Conf.Auth.SecretKey, ""),
		Secure: true,
		Region: s3Conf.Region,
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
			TLSHandshakeTimeout: s3Conf.ConnectionTimeout,
			ResponseHeaderTimeout: s3Conf.RequestTimeout,
			ExpectContinueTimeout: s3Conf.RequestTimeout,
		},
	})
}