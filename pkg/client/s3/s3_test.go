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

// hasEndpoint reports if s3 endpoint is set for tests that require it.
func hasEndpoint() bool {
	return os.Getenv("S3_HOSTNAME") != ""
}

func TestNew(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock err %s", err)
	}

	if _, got := New(mock.S3); got != nil {
		t.Errorf("transport create err %s", got)
	}
}

func TestNewErrs(t *testing.T) {
	tests := map[string]struct {
		input *secret.S3
		want  error
	}{
		"empty": {
			input: &secret.S3{},
			want:  ErrClientPrep,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, got := New(test.input); !errors.Is(got, test.want) {
				t.Errorf("unexpected err %v, want %v", got, test.want)
			}
		})
	}
}

func TestBktExistErrs(t *testing.T) {
	var (
		secrets = &secret.S3{}

		client, _ = minio.New(
			secrets.Endpoint,

			&minio.Options{
				Creds:  credentials.NewStaticV4(secrets.Key, secrets.Secret, ""),
				Secure: true,
			},
		)

		input = &Transport{
			client: client,
			bucket: secrets.Bucket,
		}
	)

	if _, got := input.client.BucketExists(context.Background(), input.bucket); got == nil {
		t.Errorf("unexpected bucket exists success")
	}
}

// Will expectedly fail either from credentials or network.
func TestUl(t *testing.T) {
	if !hasEndpoint() {
		t.Skip("s3: requires S3_HOSTNAME env var")
	}

	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock err %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create err %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatal("file write err", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	if got := transport.Ul(file.Name(), file); got != nil {
		t.Errorf("upload err %v", got)
	}
}

func TestUlErrs(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock err %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create err %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatal("file write err", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	tests := map[string]struct {
		key  string
		file *os.File
		want error
	}{
		"traverse": {
			key:  t.Name(),
			file: file,
			want: fsys.ErrPathTraverse,
		},

		"dot-dots": {
			key:  "/../etc/passwd",
			file: file,
			want: fsys.ErrPathTraverse,
		},

		"root": {
			key:  "/",
			file: file,
			want: fsys.ErrPathTraverse,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := transport.Ul(test.key, test.file); !errors.Is(got, test.want) {
				t.Errorf("unexpected err %v, want %v", got, test.want)
			}
		})
	}
}

// Will expectedly fail either from credentials or network.
func TestDl(t *testing.T) {
	if !hasEndpoint() {
		t.Skip("s3: requires S3_HOSTNAME env var")
	}

	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock err %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create err %s", err)
	}

	file, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatal("file write err", err)
	}

	defer func() {
		// lint
		_ = file.Close()
	}()

	if err := transport.Ul(file.Name(), file); err != nil {
		t.Errorf("upload err %v", err)
	}

	if err := os.Truncate(file.Name(), 0); err != nil {
		t.Fatal("file truncate err", err)
	}

	attr := &fsys.Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	if got := transport.Dl(file.Name(), attr); got != nil {
		t.Errorf("download err %v", got)
	}
}

func TestDlErrs(t *testing.T) {
	mock, err := secret.Mock(filepath.Join(t.TempDir(), t.Name()))
	if err != nil {
		t.Fatalf("secrets mock err %s", err)
	}

	transport, err := New(mock.S3)
	if err != nil {
		t.Fatalf("transport create err %s", err)
	}

	attr := &fsys.Attr{
		UID:  os.Getuid(),
		GID:  os.Getgid(),
		Mode: 0600,
	}

	tests := map[string]struct {
		input string
		want  error
	}{
		"traverse": {
			input: t.Name(),
			want:  fsys.ErrPathNotAbs,
		},

		"dot-dots": {
			input: "/../etc/passwd",
			want:  fsys.ErrPathTraverse,
		},

		"root": {
			input: "/",
			want:  fsys.ErrPathRoot,
		},

		"not-exist": {
			input: filepath.Join(t.TempDir(), "noexist", t.Name()),
			want:  fsys.ErrFileOpen,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := transport.Dl(test.input, attr); !errors.Is(got, test.want) {
				t.Errorf("unexpected download err %v, want %v", got, test.want)
			}
		})
	}
}
