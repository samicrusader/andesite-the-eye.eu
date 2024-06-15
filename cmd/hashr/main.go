package main

import (
	"github.com/nektro/go-util/util"
	"github.com/nektro/go-util/vflag"
	etc "github.com/nektro/go.etc"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/db"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/fsdb"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	etc.AppID = "andesite"
	util.Log("Initializing hashr...")

	vflag.StringVar(&idata.Config.Public, "public", "", "Public root of files")
	vflag.BoolVar(&idata.Config.Verbose, "verbose", false, "")
	vflag.BoolVar(&idata.Config.VerboseFS, "fsdb-verbose", false, "")
	vflag.StringArrayVar(&idata.Config.OffHashes, "disable-hash", []string{}, "disable hashing for dirs (one per dir)")
	vflag.IntVar(&idata.Config.ScanSimul, "scan-concurrency", runtime.NumCPU(), "number of threads to use for fsdb scan queueing")
	vflag.StringArrayVar(&idata.Config.HashExcl, "hashing-exclude", []string{}, "number of threads to use for fsdb queuing")
	etc.PreInit()
	etc.Init(&idata.Config, "./files/", db.SaveOAuth2InfoCb)

	for _, item := range idata.Config.CRootsPub {
		idata.Config.RootsPub = append(idata.Config.RootsPub, strings.SplitN(item, "=", 2))
	}
	db.Init()
	db.Upgrade()
	util.RunOnClose(func() {
		util.Log("Gracefully shutting down...")

		util.Log("Saving database to disk")
		db.Close()

		util.Log("Done!")
	})
	if len(idata.Config.Public) > 0 {
		idata.Config.Public, _ = filepath.Abs(filepath.Clean(strings.ReplaceAll(idata.Config.Public, "~", idata.HomedirPath)))
		util.DieOnError(util.Assert(util.DoesDirectoryExist(idata.Config.Public), "Public root directory does not exist. Aborting!"))
		idata.DataPathsPub["public"] = idata.Config.Public
	}
	for _, item := range idata.Config.RootsPub {
		ab, err := filepath.Abs(item[1])
		util.DieOnError(err)
		idata.DataPathsPub[item[0]] = ab
	}

	fsdb.Init(idata.DataPathsPub, "public")
}
