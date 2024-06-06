package db

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"

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
	id := DB.QueryNextID(ctFile)
	DB.Build().Ins(ctFile, id, rt, pt, sz, mt, h1, h2, h3, h4, h5, h6).Exe()
}

func DeleteFile(rt, pt string) {
	DB.Build().Del(ctFile).Wr("path", "like", pt).Exe()
}

func DropFilesFromRoot(rt string) {
	DB.Build().Del(ctFile).Wh("root", rt).Exe()
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
	return DB.Build().Se("*").Fr(ctFile)
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

func (v *File) PopulateHashes(buf []byte) bool {
	fullBuf := buf
	path := v.PathFull
	file, err := os.Open(path)
	if err != nil {
		util.LogError("fsdb:", "hasher:", "Failure reading file", path+":", err)
		return false
	}
	hMD5 := md5.New()
	hSHA1 := sha1.New()
	hSHA256 := sha256.New()
	hSHA512 := sha512.New()
	hSHA3 := sha3.New512()
	hBLAKE2b, _ := blake2b.New512(make([]byte, 0))
	eof := false
	for !eof {
		var n int
		n, err = file.Read(buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			eof = true
		} else if err != nil {
			util.LogError("fsdb:", "hasher:", "Failure reading file", path+":", err)
			return false
		}
		buf = buf[:n]
		for _, hash := range idata.Hashes {
			switch hash {
			case "MD5":
				hMD5.Write(buf)
			case "SHA1":
				hSHA1.Write(buf)
			case "SHA256":
				hSHA256.Write(buf)
			case "SHA512":
				hSHA512.Write(buf)
			case "SHA3_512":
				hSHA3.Write(buf)
			case "BLAKE2b_512":
				hBLAKE2b.Write(buf)
			}
		}
		buf = fullBuf
	}
	for _, hash := range idata.Hashes {
		switch hash {
		case "MD5":
			v.MD5 = hex.EncodeToString(hMD5.Sum(nil))
		case "SHA1":
			v.SHA1 = hex.EncodeToString(hSHA1.Sum(nil))
		case "SHA256":
			v.SHA256 = hex.EncodeToString(hSHA256.Sum(nil))
		case "SHA512":
			v.SHA512 = hex.EncodeToString(hSHA512.Sum(nil))
		case "SHA3_512":
			v.SHA3 = hex.EncodeToString(hSHA3.Sum(nil))
		case "BLAKE2b_512":
			v.BLAKE2b = hex.EncodeToString(hBLAKE2b.Sum(nil))
		}
	}
	return true
}

func (v *File) SetSize(x int64) {
	v.Size = x
	n := strconv.FormatInt(x, 10)
	Up(v, DB, ctFile, "size", n)
}

func (v *File) SetModTime(x int64) {
	v.ModTime = x
	n := strconv.FormatInt(x, 10)
	Up(v, DB, ctFile, "mod_time", n)
}
