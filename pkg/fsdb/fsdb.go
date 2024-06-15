package fsdb

import (
	"github.com/samicrusader/andesite-the-eye.eu/pkg/db"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/karrick/godirwalk"
	"github.com/nektro/go-util/util"
)

const bufSize = 1 << 21

var syncer = int64(1 << 62)

type Job struct {
	f *db.File
}

func Init(mp map[string]string, rt string) {
	bd, ok := mp[rt]
	if !ok || idata.Config.NoAutoScan {
		return
	}
	jobs := make(chan Job)

	for i := 0; i < idata.Config.ScanSimul; i++ {
		go worker(jobs)
	}

	util.Log("fsdb:", "init: walking directory", bd)
	godirwalk.Walk(bd, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			fp, _ := filepath.Abs(osPathname)
			fi, _ := os.Lstat(fp)

			// Remove base directory from path
			relpath := "/" + rt + strings.TrimPrefix(fp, bd)

			for _, item := range idata.Config.HashExcl {
				if strings.TrimSuffix(item, "/") == fp {
					util.Log("fsdb:", "skipping", fp, "due to exclude")
					return godirwalk.SkipThis
				}
			}

			// Remove dotfiles and ignore if directory
			if strings.HasSuffix(relpath, "/"+rt+"/.") || fi.IsDir() {
				return nil
			}

			// Symlink detection
			if fi.Mode()&os.ModeSymlink != 0 {
				sympath, _ := filepath.EvalSymlinks(fp)

				if idata.Config.VerboseFS {
					util.Log("fsdb:", "walk: hit a symlink:", fp)
				}

				if sympath == "" {
					util.LogError("fsdb:", "walk:", "symlink", fp, "is pointing to a non-existing file")
					return godirwalk.SkipThis
				}

				// Query symlink
				s, err := os.Lstat(sympath)
				if err != nil {
					util.LogError("fsdb:", "walk/symlink:", err)
					return nil
				}
				if s.IsDir() {
					return nil
				}
			}
			f := &db.File{
				0,
				rt,
				relpath, fp,
				fi.Size(), "",
				fi.ModTime().UTC().Unix(), "",
				"", "", "", "", "", "",
			}
			jobs <- Job{f}
			atomic.AddInt64(&syncer, 1)
			return nil
		},
		Unsorted:            true,
		FollowSymbolicLinks: true,
	})
	util.Log("fsdb:", "init: done walking directory", bd)
}

func worker(jobs <-chan Job) {
	buf := make([]byte, bufSize)
	for job := range jobs {
		insertFile(job, buf)
		atomic.AddInt64(&syncer, -1)
	}
}

func insertFile(job Job, buf []byte) {
	f := job.f
	// Check against old DB entry
	oldentry, ok := db.File{}.ByPath(f.Path)
	if ok {
		// File exists and modified time has not changed
		if oldentry.ModTime == f.ModTime {
			if idata.Config.VerboseFS {
				util.Log("fsdb:", "skipped:", f.Path)
			}
			return
		}

		if idata.Config.VerboseFS {
			util.Log("fsdb:", "processing:", oldentry.Path)
		}

		// File exists but changed
		oldentry.PathFull = f.PathFull
		ok := oldentry.PopulateHashes(buf)
		if !ok {
			return
		}
		oldentry.SetSize(f.Size)
		oldentry.SetModTime(f.ModTime)
		db.DeleteFile(oldentry.Root, oldentry.Path)
		db.CreateFile(oldentry.Root, oldentry.Path, oldentry.Size, oldentry.ModTime, oldentry.MD5, oldentry.SHA1, oldentry.SHA256, oldentry.SHA512, oldentry.SHA3, oldentry.BLAKE2b)
		if idata.Config.VerboseFS {
			util.Log("fsdb:", "updated:", oldentry.Path)
		}
		return
	} else {
		// File does not exist, add
		ok := f.PopulateHashes(buf)
		if !ok {
			return
		}
		db.CreateFile(f.Root, f.Path, f.Size, f.ModTime, f.MD5, f.SHA1, f.SHA256, f.SHA512, f.SHA3, f.BLAKE2b)
		if idata.Config.VerboseFS {
			util.Log("fsdb:", "added:", f.Path)
		}
		return
	}
}

func DeInit(mp map[string]string, rt string) {
	_, ok := mp[rt]
	if !ok {
		return
	}
	db.DropFilesFromRoot(rt)
	util.Log("fsdb:", rt+":", "removed.")
}
