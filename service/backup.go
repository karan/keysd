package service

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/keys-pub/keys/keyring/backup"
	"github.com/pkg/errors"
)

// BackupExport ...
func (s *service) BackupExport(ctx context.Context, req *BackupExportRequest) (*BackupExportResponse, error) {
	ur, err := url.Parse(req.URI)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse URI")
	}
	if ur.Scheme != "file" {
		return nil, errors.Errorf("URI should be in the format file://dir")
	}
	if req.Password == "" {
		return nil, errors.Errorf("no password specified")
	}

	dir := ur.Path
	if !filepath.IsAbs(dir) {
		return nil, errors.Errorf("absolute path required: %s", dir)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, errors.Errorf("dir doesn't exist: %s", dir)
	}

	path, err := backup.ExportToDirectory(s.ks.Keyring(), dir, req.Password, nil)
	if err != nil {
		return nil, err
	}

	return &BackupExportResponse{
		URI: fmt.Sprintf("file://%s", path),
	}, nil
}

// BackupImport ...
func (s *service) BackupImport(ctx context.Context, req *BackupImportRequest) (*BackupImportResponse, error) {
	ur, err := url.Parse(req.URI)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse URI")
	}
	if ur.Scheme != "file" {
		return nil, errors.Errorf("URI should be in the format file://dir")
	}
	if req.Password == "" {
		return nil, errors.Errorf("no password specified")
	}

	path := ur.Path
	if !filepath.IsAbs(path) {
		return nil, errors.Errorf("absolute path required: %s", path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.Errorf("path doesn't exist: %s", path)
	}

	if err := backup.ImportFromFile(s.ks.Keyring(), path, req.Password); err != nil {
		return nil, err
	}

	return &BackupImportResponse{}, nil
}
