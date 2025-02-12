// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package hit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"

	"github.com/defended-net/malwatch/pkg/boot/env/re"
	"github.com/defended-net/malwatch/pkg/db/orm"
	"github.com/defended-net/malwatch/pkg/fsys"
)

// Paths represents path:hits.
type Paths map[string][]*Meta

// Meta represents hit meta.
type Meta struct {
	Time   time.Time  `json:"time"`
	Rules  []string   `json:"rules"`
	Status string     `json:"status"`
	Attr   *fsys.Attr `json:"attr"`
	Acts   []string   `json:"actions"`
}

// History represents target hit history.
type History struct {
	Target string `json:"target"`
	Paths  Paths  `json:"paths"`
}

const (
	bucket = "hits"
	tzFmt  = "2006-01-02 15:04:05"
)

// NewMeta returns new hit meta.
func NewMeta(attr *fsys.Attr, matches []string, verbs ...string) *Meta {
	return &Meta{
		Time:  time.Now(),
		Rules: matches,
		Attr:  attr,
		Acts:  verbs,
	}
}

// SelectAll returns all hits.
func SelectAll(db *bbolt.DB) ([]*History, error) {
	if db == nil {
		return nil, orm.ErrDbNotLoaded
	}

	hits, err := orm.GetAll(db, bucket)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", orm.ErrBktIter, err)
	}

	var histories []*History
	for _, hit := range hits {
		var paths Paths
		if err := json.Unmarshal(hit.Val, &paths); err != nil {
			return nil, fmt.Errorf("%w, %v", orm.ErrMarshal, err)
		}

		histories = append(histories, &History{
			Target: hit.Key,
			Paths:  paths,
		})
	}

	return histories, nil
}

// SelectTarget returns a target's hits.
func SelectTarget(db *bbolt.DB, target string) (Paths, error) {
	if db == nil {
		return nil, orm.ErrDbNotLoaded
	}

	paths := Paths{}

	hits, err := orm.Get(db, bucket, target)
	if err != nil {
		return paths, nil
	}

	if hits.Val == nil {
		return paths, nil
	}

	if err := json.Unmarshal(hits.Val, &paths); err != nil {
		return nil, fmt.Errorf("%w, %v", orm.ErrMarshal, err)
	}

	return paths, nil
}

// SelectLast returns a path's most recent hit.
func SelectLast(db *bbolt.DB, path string) (*Meta, error) {
	if db == nil {
		return nil, orm.ErrDbNotLoaded
	}

	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("%w, %v", fsys.ErrPathNotAbs, path)
	}

	target := re.Target(path)

	history, err := SelectTarget(db, target)
	if err != nil {
		return nil, fmt.Errorf("%w, %v, %v", orm.ErrBktIter, err, target)
	}

	if metas, ok := history[path]; ok && len(metas) > 0 {
		return metas[len(metas)-1], nil
	}

	return &Meta{}, nil
}

// Save saves hits.
func (hits *History) Save(db *bbolt.DB) error {
	if db == nil || len(hits.Paths) == 0 {
		return nil
	}

	history, err := SelectTarget(db, hits.Target)
	if err != nil {
		return err
	}

	for path, metas := range hits.Paths {
		history[path] = append(history[path], metas...)
	}

	return orm.Put(db, bucket, hits.Target, history)
}

// DelTarget deletes target hits. Entire key is removed.
func DelTarget(db *bbolt.DB, target string) error {
	if db == nil {
		return orm.ErrDbNotLoaded
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))

		return bucket.Delete([]byte(target))
	})
}

// DelPath deletes path hits.
func DelPath(db *bbolt.DB, target string, path string) error {
	if db == nil {
		return orm.ErrDbNotLoaded
	}

	hits, err := SelectTarget(db, target)
	if err != nil {
		return fmt.Errorf("%w, %v", orm.ErrBktIter, target)
	}

	delete(hits, path)

	return orm.Put(db, bucket, target, hits)
}

// Restore restores a quarantined hit.
func (meta *Meta) Restore(quarantineDir string, dst string) error {
	src := filepath.Join(quarantineDir, filepath.Dir(dst), meta.Status)

	if err := fsys.HasDotDots(src, dst); err != nil {
		return err
	}

	attr := &fsys.Attr{
		UID:  meta.Attr.UID,
		GID:  meta.Attr.GID,
		Mode: meta.Attr.Mode,
	}

	return fsys.Mv(src, dst, attr)
}

// ToSlice returns string slices from given history.
func (hits *History) ToSlice() [][]string {
	rows := [][]string{}

	for path, metas := range hits.Paths {
		for _, meta := range metas {
			rows = append(rows, []string{
				path,
				strings.Join(meta.Rules, "\n"),
				strings.Join(meta.Acts, ","),
			})
		}
	}

	return rows
}

// ToSlice returns string slices from given paths.
func (hits Paths) ToSlice() [][]string {
	var rows [][]string

	for path, metas := range hits {
		for _, meta := range metas {
			rows = append(rows, meta.ToSlice(path))
		}
	}

	return rows
}

// ToSlice returns a string slice from given meta and path.
func (meta *Meta) ToSlice(path string) []string {
	if meta.Attr == nil {
		return []string{}
	}

	return []string{
		path,
		meta.Time.UTC().Format(tzFmt),
		strings.Join(meta.Rules, "\n"),
		meta.Attr.MTime.Format(tzFmt),
		meta.Attr.CTime.Format(tzFmt),
		strconv.FormatUint(uint64(meta.Attr.UID), 10),
		strconv.FormatUint(uint64(meta.Attr.GID), 10),
		meta.Status,
		strings.Join(meta.Acts, ","),
	}
}
