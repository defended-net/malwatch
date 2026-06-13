// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package s3

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"golang.org/x/sys/unix"

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

// New returns new transport.
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

// Dl downloads a file. Perms and owner from given attr.
func (transport *Transport) Dl(path string, attr *fsys.Attr) error {
	if err := fsys.HasDotDots(path); err != nil {
		return err
	}

	fd, _, err := fsys.OpenFile(path, unix.O_WRONLY|unix.O_CREAT|unix.O_TRUNC)
	if err != nil {
		return err
	}

	// #nosec G115 -- os file desc.
	file := os.NewFile(uintptr(fd), path)
	defer fsys.Close(file)

	if err := unix.Fchmod(fd, 0600); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChmod, err, path)
	}

	obj, err := transport.client.GetObject(
		context.Background(),
		transport.bucket,
		path,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("%w, %v, %v", ErrObjGet, err, path)
	}

	if _, err := io.Copy(file, obj); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrObjGet, err, path)
	}

	if err := unix.Fchmod(fd, uint32(attr.Mode)); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChmod, err, path)
	}

	if err := unix.Fchown(fd, attr.UID, attr.GID); err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrChown, err, path)
	}

	slog.Info("download complete", "path", path)

	return nil
}

// Ul uploads file from given path.
func (transport *Transport) Ul(key string, file *os.File) error {
	if err := fsys.HasDotDots(key); err != nil {
		return fmt.Errorf("%w, %v", fsys.ErrPathTraverse, key)
	}

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("%w, %v, %v", fsys.ErrStat, err, key)
	}

	if _, err = transport.client.PutObject(
		context.Background(),
		transport.bucket,
		key,
		file,
		info.Size(),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	); err != nil {
		return fmt.Errorf("%w, %v, %v", ErrObjPut, err, key)
	}

	slog.Info("upload complete", "path", key)

	return nil
}
