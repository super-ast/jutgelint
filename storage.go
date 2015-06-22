/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// Length of the random hexadecimal ids. At least 4.
	idSize = 8
	// Number of times to try getting an unused random id
	randTries = 10
)

var (
	// ErrReviewNotFound means that we could not find the requested review
	ErrReviewNotFound = errors.New("review could not be found")
	// ErrNoUnusedIDFound means that we could not find an unused ID to
	// allocate to a new review
	ErrNoUnusedIDFound = errors.New("gave up trying to find an unused random id")
)

// A Review represents the review's content and information
type Review interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
	ModTime() time.Time
}

// ID is the binary representation of the identifier for a review
type ID [idSize / 2]byte

// IDFromString parses a hexadecimal string into an ID. Returns the ID and an
// error, if any.
func IDFromString(hexID string) (id ID, err error) {
	if len(hexID) != idSize {
		return id, fmt.Errorf("invalid id at %s", hexID)
	}
	b, err := hex.DecodeString(hexID)
	if err != nil || len(b) != idSize/2 {
		return id, fmt.Errorf("invalid id at %s", hexID)
	}
	copy(id[:], b)
	return id, nil
}

func (id ID) String() string {
	return hex.EncodeToString(id[:])
}

// A Store represents a database holding multiple reviews identified by their
// ids
type Store interface {
	// Get the review known by the given ID and an error, if any.
	Get(id ID) (Review, error)

	// Put a new review given its content. Will return the ID assigned to
	// the new review and an error, if any.
	Put(content []byte) (ID, error)
}

func randomID(available func(ID) bool) (ID, error) {
	var id ID
	for try := 0; try < randTries; try++ {
		if _, err := rand.Read(id[:]); err != nil {
			continue
		}
		if available(id) {
			return id, nil
		}
	}
	return id, ErrNoUnusedIDFound
}

type FileStore struct {
	sync.RWMutex
	cache map[ID]fileCache
	dir   string
}

type FileReview struct {
	file  *os.File
	cache *fileCache
}

type fileCache struct {
	path    string
	modTime time.Time
}

func (c FileReview) Read(p []byte) (n int, err error) {
	return c.file.Read(p)
}

func (c FileReview) ReadAt(p []byte, off int64) (n int, err error) {
	return c.file.ReadAt(p, off)
}

func (c FileReview) Seek(offset int64, whence int) (int64, error) {
	return c.file.Seek(offset, whence)
}

func (c FileReview) Close() error {
	return c.file.Close()
}

func (c FileReview) ModTime() time.Time {
	return c.cache.modTime
}

func NewFileStore(dir string) (*FileStore, error) {
	if err := setupTopDir(dir); err != nil {
		return nil, err
	}
	s := new(FileStore)
	s.dir = dir
	s.cache = make(map[ID]fileCache)

	insert := func(id ID, path string, modTime time.Time) error {
		cached := fileCache{
			path:    path,
			modTime: modTime,
		}
		s.cache[id] = cached
		return nil
	}
	if err := setupSubdirs(s.dir, fileRecover(insert, s)); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStore) Get(id ID) (Review, error) {
	s.RLock()
	defer s.RUnlock()
	cached, e := s.cache[id]
	if !e {
		return nil, ErrReviewNotFound
	}
	f, err := os.Open(cached.path)
	if err != nil {
		return nil, err
	}
	return FileReview{file: f, cache: &cached}, nil
}

func writeNewFile(filename string, data []byte) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func (s *FileStore) Put(content []byte) (ID, error) {
	available := func(id ID) bool {
		_, e := s.cache[id]
		return !e
	}
	s.Lock()
	defer s.Unlock()
	id, err := randomID(available)
	if err != nil {
		return id, err
	}
	reviewPath := pathFromID(id)
	if err = writeNewFile(reviewPath, content); err != nil {
		return id, err
	}
	s.cache[id] = fileCache{
		path:    reviewPath,
		modTime: time.Now(),
	}
	return id, nil
}

func pathFromID(id ID) string {
	hexID := id.String()
	return path.Join(hexID[:2], hexID[2:])
}

func idFromPath(path string) (ID, error) {
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) != 2 {
		return ID{}, fmt.Errorf("invalid number of directories at %s", path)
	}
	if len(parts[0]) != 2 {
		return ID{}, fmt.Errorf("invalid directory name length at %s", path)
	}
	hexID := parts[0] + parts[1]
	return IDFromString(hexID)
}

type fileInsert func(id ID, path string, modTime time.Time) error

func fileRecover(insert fileInsert, s Store) filepath.WalkFunc {
	return func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() {
			return err
		}
		id, err := idFromPath(path)
		if err != nil {
			return err
		}
		modTime := fileInfo.ModTime()
		if err := insert(id, path, modTime); err != nil {
			return err
		}
		return nil
	}
}

func setupTopDir(topdir string) error {
	if err := os.MkdirAll(topdir, 0700); err != nil {
		return err
	}
	return os.Chdir(topdir)
}

func setupSubdirs(topdir string, rec filepath.WalkFunc) error {
	for i := 0; i < 256; i++ {
		if err := setupSubdir(topdir, rec, byte(i)); err != nil {
			return err
		}
	}
	return nil
}

func setupSubdir(topdir string, rec filepath.WalkFunc, h byte) error {
	dir := hex.EncodeToString([]byte{h})
	if stat, err := os.Stat(dir); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("%s/%s exists but is not a directory", topdir, dir)
		}
		if err := filepath.Walk(dir, rec); err != nil {
			return fmt.Errorf("cannot recover data directory %s/%s: %s", topdir, dir, err)
		}
	} else if err := os.Mkdir(dir, 0700); err != nil {
		return fmt.Errorf("cannot create data directory %s/%s: %s", topdir, dir, err)
	}
	return nil
}
