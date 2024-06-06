package db

import (
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"

	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"
)

type File struct {
	ID       int64  `json:"id"`
	Root     string `json:"root" dbsorm:"1"`
	Path     string `json:"path" dbsorm:"1"`
	PathFull string
	Size     int64  `json:"size" dbsorm:"1"`
	SizeS    string `json:"html_size"`
	ModTime  int64  `json:"mod_time" dbsorm:"1"`
	ModTimeS string `json:"html_modtime"`
	MD5      string `json:"hash_md5" dbsorm:"1"`
	SHA1     string `json:"hash_sha1" dbsorm:"1"`
	SHA256   string `json:"hash_sha256" dbsorm:"1"`
	SHA512   string `json:"hash_sha512" dbsorm:"1"`
	SHA3     string `json:"hash_sha3" dbsorm:"1"`
	BLAKE2b  string `json:"hash_blake2b" dbsorm:"1"`
}

func CreateFile(rt, pt string, sz, mt int64, h1, h2, h3, h4, h5, h6 string) {
	dbstorage.InsertsLock.Lock()
	defer dbstorage.InsertsLock.Unlock()
	//
	id := db.QueryNextID(ctFile)
	db.Build().Ins(ctFile, id, rt, pt, sz, mt, h1, h2, h3, h4, h5, h6).Exe()
}

func DropFilesFromRoot(rt string) {
	db.Build().Del(ctFile).Wh("root", rt).Exe()
}

// Scan implements dbstorage.Scannable
func (v File) Scan(rows *sql.Rows) dbstorage.Scannable {
	rows.Scan(&v.ID, &v.Root, &v.Path, &v.Size, &v.ModTime, &v.MD5, &v.SHA1, &v.SHA256, &v.SHA512, &v.SHA3, &v.BLAKE2b)
	v.SizeS = util.ByteCountIEC(v.Size)
	v.ModTimeS = time.Unix(v.ModTime, -1).UTC().Format(time.RFC822)
	return &v
}

func (File) ScanAll(q dbstorage.QueryBuilder) []*File {
	arr := dbstorage.ScanAll(q, File{})
	res := []*File{}
	for _, item := range arr {
		o, ok := item.(*File)
		if !ok {
			continue
		}
		res = append(res, o)
	}
	return res
}

func (v *File) i() string {
	return strconv.FormatInt(v.ID, 10)
}

func (File) b() dbstorage.QueryBuilder {
	return db.Build().Se("*").Fr(ctFile)
}

func (File) All() []*File {
	return File{}.ScanAll(File{}.b())
}

//
// searchers
//

func (File) ByPath(path string) (*File, bool) {
	ur, ok := dbstorage.ScanFirst(File{}.b().Wh("path", path), File{}).(*File)
	return ur, ok
}

//
// modifiers
//

func (v *File) PopulateHashes(doUp bool) {
	for _, item := range idata.Hashes {
		v.setHash(item, hash(item, v.PathFull), doUp)
	}
}

func (v *File) setHash(alg, hv string, doUp bool) {
	switch alg {
	case "MD5":
		v.MD5 = hv
	case "SHA1":
		v.SHA1 = hv
	case "SHA256":
		v.SHA256 = hv
	case "SHA512":
		v.SHA512 = hv
	case "SHA3_512":
		v.SHA3 = hv
	case "BLAKE2b_512":
		v.BLAKE2b = hv
	}
	if doUp && stringsu.Contains(idata.Hashes, alg) {
		hk := strings.ToLower(strings.TrimSuffix(alg, "_512"))
		db.Build().Up(ctFile, "hash_"+hk, hv).Wh("id", v.i()).Exe()
	}
}

func hash(algo string, pathS string) string {
	f, _ := os.Open(pathS)
	defer f.Close()
	return util.HashStream(algo, f)
}

func (v *File) SetSize(x int64) {
	v.Size = x
	n := strconv.FormatInt(x, 10)
	Up(v, db, ctFile, "size", n)
}

func (v *File) SetModTime(x int64) {
	v.ModTime = x
	n := strconv.FormatInt(x, 10)
	Up(v, db, ctFile, "mod_time", n)
}
