package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	azip "github.com/pierrre/archivefile/zip"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func about() {
	fmt.Printf("This is burry in version %s\n", VERSION)
}

// lookupst returns the storage target index
// based on a name (or -1 if not known)
func lookupst(name string) int {
	switch strings.ToLower(name) {
	case STORAGE_TARGET_TTY:
		return 0
	case STORAGE_TARGET_LOCAL:
		return 1
	case STORAGE_TARGET_S3:
		return 2
	case STORAGE_TARGET_MINIO:
		return 3
	default:
		return -1
	}
}

// reapsimple reaps a node at a path.
// note that the actual processing is
// determined by the storage target
func reapsimple(path string, val string) {
	for _, prefix := range blacklist {
		if len(prefix) > 0 && strings.HasPrefix(path, prefix) {
			log.WithFields(log.Fields{"func": "reapsimple"}).Debugf("Skipping blacklisted %s (prefix: %s)", path, prefix)
			return
		}
	}
	stidx := lookupst(brf.StorageTarget)
	switch {
	case stidx == 0: // TTY
		log.WithFields(log.Fields{"func": "reapsimple"}).Info(fmt.Sprintf("%s", path))
		log.WithFields(log.Fields{"func": "reapsimple"}).Debug(fmt.Sprintf("%v", val))
	case stidx >= 1: // some kind of actual storage
		store(path, val)
	default:
		log.WithFields(log.Fields{"func": "reapsimple"}).Fatal(fmt.Sprintf("Storage target %s unknown or not yet supported", brf.StorageTarget))
	}
}

// store stores value val at path path
// in the local filesystem
func store(path string, val string) {
	cwd, _ := os.Getwd()
	fpath := ""
	if path == "/" {
		log.WithFields(log.Fields{"func": "store"}).Info(fmt.Sprintf("Rewriting root"))
		fpath, _ = filepath.Abs(filepath.Join(cwd, based))
	} else {
		// escape ":" in the path so that we have no issues storing it in the filesystem:
		fpath, _ = filepath.Abs(filepath.Join(cwd, based, strings.Replace(path, ":", "BURRY_ESC_COLON", -1)))
	}
	if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
		log.WithFields(log.Fields{"func": "store"}).Error(fmt.Sprintf("%s", err))
		return
	} else {
		cpath, _ := filepath.Abs(filepath.Join(fpath, CONTENT_FILE))
		if c, cerr := os.Create(cpath); cerr != nil {
			log.WithFields(log.Fields{"func": "store"}).Error(fmt.Sprintf("%s", cerr))
		} else {
			defer c.Close()
			if nbytes, err := c.WriteString(val); err != nil {
				log.WithFields(log.Fields{"func": "store"}).Error(fmt.Sprintf("%s", err))
			} else {
				log.WithFields(log.Fields{"func": "store"}).Debug(fmt.Sprintf("Stored %s in %s with %d bytes", path, fpath, nbytes))
			}
		}
	}
}

// arch creates a ZIP archive of snapshot that store() has generated
func arch() string {
	defer func() {
		_ = os.RemoveAll(based)
	}()
	cwd, _ := os.Getwd()
	opath := filepath.Join(cwd, based+".zip")
	ipath := filepath.Join(cwd, based, "/")
	progress := func(apath string) {
		log.WithFields(log.Fields{"func": "arch"}).Debug(fmt.Sprintf("%s", apath))
	}
	// add metadata ot the archive:
	addmeta(ipath)
	if err := azip.ArchiveFile(ipath, opath, progress); err != nil {
		log.WithFields(log.Fields{"func": "arch"}).Panic(fmt.Sprintf("%s", err))
	} else {
		log.WithFields(log.Fields{"func": "arch"}).Debug(fmt.Sprintf("Backup available in %s", opath))
	}
	return opath
}

// unarch creates a directory with contents of the snapshot
// based on the ZIP archive from an earlier backup operation
func unarch(localarch string) string {
	cwd, _ := os.Getwd()
	ipath := localarch
	opath := cwd
	progress := func(apath string) {
		log.WithFields(log.Fields{"func": "unarch"}).Debug(fmt.Sprintf("%s", apath))
	}
	if err := azip.UnarchiveFile(ipath, opath, progress); err != nil {
		log.WithFields(log.Fields{"func": "unarch"}).Panic(fmt.Sprintf("%s", err))
	} else {
		log.WithFields(log.Fields{"func": "unarch"}).Debug(fmt.Sprintf("Backup restored in %s", opath))
	}
	return filepath.Join(cwd, based, "/")
}

// readc reads a file from the specified path and
// returns its content as a byte slice
func readc(path string) ([]byte, error) {
	c := []byte{}
	if _, ferr := os.Stat(path); ferr != nil {
		return c, ferr
	} else { // content file exists
		if c, rerr := ioutil.ReadFile(path); rerr != nil {
			return c, rerr
		} else {
			return c, nil
		}
	}
}

// MD5All reads all the files in the file tree rooted at root and returns
// from file path to the MD5 sum of the file's contents.  If the directory walk
// fails or any read operation fails, MD5All returns an error.
func MD5All(root string) ([]byte, error) {
	h := md5.New()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// add names to the hash as well, but remove the root
		io.WriteString(h, strings.TrimPrefix(path, root))

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			if _, err := io.Copy(h, f); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		return []byte{0x00}, err
	}
	return h.Sum(nil), nil
}

func hasArchiveChanged() bool {
	cwd, _ := os.Getwd()
	ipath := filepath.Join(cwd, based, "/")
	md5, err := MD5All(ipath)
	if err != nil {
		log.WithFields(log.Fields{"func": "hasArchiveChanged"}).Fatalf("Could not calculate MD5 hash of %s (%v)", ipath, err)
	}
	log.WithFields(log.Fields{"func": "hasArchiveChanged"}).Debugf("Old checksum: %x, new checksum: %x", checksum, md5)
	result := !bytes.Equal(md5, checksum)
	checksum = md5
	return result
}
