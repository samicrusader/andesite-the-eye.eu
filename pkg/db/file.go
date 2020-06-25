package db

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"
)

type File struct {
	ID       int64 `json:"id"`
	IDS      string
	Root     string `json:"root" sqlite:"text"`
	Path     string `json:"path" sqlite:"text"`
	Size     int64  `json:"size" sqlite:"int"`
	SizeS    string `json:"html_size"`
	ModTime  int64  `json:"mod_time" sqlite:"int"`
	ModTimeS string `json:"html_modtime"`
	MD5      string `json:"hash_md5" sqlite:"text"`
	SHA1     string `json:"hash_sha1" sqlite:"text"`
	SHA256   string `json:"hash_sha256" sqlite:"text"`
	SHA512   string `json:"hash_sha512" sqlite:"text"`
	SHA3     string `json:"hash_sha3" sqlite:"text"`
	BLAKE2b  string `json:"hash_blake2b" sqlite:"text"`
}

// Scan implements dbstorage.Scannable
func (v File) Scan(rows *sql.Rows) dbstorage.Scannable {
	rows.Scan(&v.ID, &v.Root, &v.Path, &v.Size, &v.ModTime, &v.MD5, &v.SHA1, &v.SHA256, &v.SHA512, &v.SHA3, &v.BLAKE2b)
	v.IDS = strconv.FormatInt(v.ID, 10)
	v.SizeS = util.ByteCountIEC(v.Size)
	v.ModTimeS = time.Unix(v.ModTime, -1).UTC().String()[:19]
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

func (File) b() dbstorage.QueryBuilder {
	return FS.Build().Se("*").Fr(ctFile)
}

func (File) All() []*File {
	return File{}.ScanAll(File{}.b())
}

//
// searchers
//

//
// modifiers
//
