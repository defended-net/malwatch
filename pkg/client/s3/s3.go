// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package s3

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Scheme stores the uri scheme.
const Scheme = "s3://"

// Transport represents the transport.
type Transport struct {
	client *minio.Client
	bucket string
}

// New returns a new transport.
func New(secrets *secret.S3) (*Transport, error) {
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(secrets.Key, secrets.Secret, ""),
		Secure: true,
	}

	client, err := minio.New(secrets.Endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrClientPrep, err)
	}

	ok, err := client.BucketExists(context.Background(), secrets.Bucket)
	if err == nil && !ok {
		if err := client.MakeBucket(
			context.Background(),
			secrets.Bucket,
			minio.MakeBucketOptions{
				Region: secrets.Region,
			}); err != nil {
			return nil, fmt.Errorf("%w, %v", ErrBktAdd, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrBktLookup, err)
	}

	return &Transport{
		client: client,
		bucket: secrets.Bucket,
	}, nil
}

// Dl downloads a file. Permissions and ownership are adjusted based on given attributes.
func (transport *Transport) Dl(path string, attr *fsys.Attr) error {
	slog.Info("downloading", "path", path)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, attr.Mode)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileOpen, err, path)
	}
	defer file.Close()

	if err := transport.client.FGetObject(
		context.Background(),
		transport.bucket,
		path,
		path,
		minio.GetObjectOptions{},
	); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrObjGet, err, path)
	}

	slog.Info("download complete", "path", file.Name())

	if err := os.Chmod(file.Name(), attr.Mode); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChmod, err, path)
	}

	if err := os.Chown(file.Name(), attr.UID, attr.GID); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChown, err, path)
	}

	return nil
}

// Ul uploads a file.
func (transport *Transport) Ul(path string) error {
	slog.Info("uploading", "path", path)

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrFileOpen, err, path)
	}
	defer file.Close()

	if _, err := transport.client.FPutObject(
		context.Background(),
		transport.bucket,
		path,
		path,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		}); err != nil {
		return err
	}

	slog.Info("upload complete", "path", file.Name())

	return nil
}
