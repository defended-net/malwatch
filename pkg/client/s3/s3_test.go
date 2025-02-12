// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package s3

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/defended-net/malwatch/pkg/boot/env/cfg/secret"
	"github.com/defended-net/malwatch/pkg/fsys"
)

func TestNew(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock error: %s", err)
	}

	if _, err = New(mock.S3); err != nil {
		t.Errorf("transport create error: %s", err)
	}
}

func TestNewClientErrs(t *testing.T) {
	secrets := &secret.S3{
		Key:      "",
		Secret:   "",
		Endpoint: "",
		Bucket:   "",
		Region:   "",
	}

	if _, err := New(secrets); err == nil {
		t.Errorf("unexpected new client success, wanted err")
	}
}

func TestBktExistErrs(t *testing.T) {
	secrets := &secret.S3{
		Key:      "",
		Secret:   "",
		Endpoint: "",
		Bucket:   "",
	}

	minioClient, _ := minio.New(secrets.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(secrets.Key, secrets.Secret, ""),
		Secure: true,
	})

	transport := &Transport{
		client: minioClient,
		bucket: secrets.Bucket,
	}

	if _, err := transport.client.BucketExists(context.Background(), transport.bucket); err == nil {
		t.Errorf("expected error, got nil")
	}
}

// Will expectedly fail either from credentials or network.
func TestUl(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock error: %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create error: %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatal("file create error:", err)
	}
	defer file.Close()

	if err := transport.Ul(file.Name()); err != nil {
		t.Errorf("upload error: %v", err)
	}
}

// Will expectedly fail either from credentials or network.
func TestDl(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock error: %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create error: %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatal("file create error:", err)
	}
	defer file.Close()

	if err := transport.Ul(file.Name()); err != nil {
		t.Errorf("upload error: %v", err)
	}

	if err := os.Remove(file.Name()); err != nil {
		t.Fatal("file remove error:", err)
	}

	attr := &fsys.Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	if err := transport.Dl(file.Name(), attr); err != nil {
		t.Errorf("dload error: %v", err)
	}
}

func TestDlErrs(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock error: %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create error: %s", err)
	}

	attr := &fsys.Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	if got := transport.Dl(filepath.Join(t.TempDir(), t.Name()), attr); !errors.Is(got, ErrObjGet) {
		t.Errorf("unexpected download error: %v, want %v", got, ErrObjGet)
	}
}

func TestUlErrs(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock error: %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Errorf("transport create error: %s", err)
	}

	if got := transport.Ul(t.Name()); !errors.Is(got, fsys.ErrFileOpen) {
		t.Errorf("unexpected upload error: %v, want %v", got, fsys.ErrFileOpen)
	}
}
