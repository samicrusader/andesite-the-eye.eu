package db

import (
	"database/sql"
	"strconv"

	dbstorage "github.com/nektro/go.dbstorage"
)

type UserAccess struct {
	ID   int64  `json:"id"`
	User int64  `json:"user" sqlite:"int"`
	Path string `json:"path" sqlite:"text"`
}

// Scan implements dbstorage.Scannable
func (v UserAccess) Scan(rows *sql.Rows) dbstorage.Scannable {
	rows.Scan(&v.ID, &v.User, &v.Path)
	return &v
}

func (UserAccess) ScanAll(q dbstorage.QueryBuilder) []*UserAccess {
	arr := dbstorage.ScanAll(q, UserAccess{})
	res := []*UserAccess{}
	for _, item := range arr {
		o, ok := item.(*UserAccess)
		if !ok {
			continue
		}
		res = append(res, o)
	}
	return res
}

func (v *UserAccess) i() string {
	return strconv.FormatInt(v.ID, 10)
}

func (UserAccess) b() dbstorage.QueryBuilder {
	return DB.Build().Se("*").Fr(ctUserAccess)
}

func (UserAccess) All() []*UserAccess {
	return UserAccess{}.ScanAll(UserAccess{}.b())
}

//
// searchers
//

func (UserAccess) ByUser(user *User) []*UserAccess {
	return UserAccess{}.ScanAll(UserAccess{}.b().Wh("user", user.i()))
}

//
// modifiers
//
