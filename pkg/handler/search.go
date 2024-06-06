package handler

import (
	"net/http"
	"strings"

	"github.com/samicrusader/andesite-the-eye.eu/pkg/config"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/db"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"

	etc "github.com/nektro/go.etc"
	oauth2 "github.com/nektro/go.oauth2"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	_, user, err := ApiBootstrap(r, w, []string{http.MethodGet}, config.GlobalSearchOff, config.GlobalSearchOff, true)
	if err != nil {
		return
	}
	etc.WriteHandlebarsFile(r, w, "/search.hbs", map[string]interface{}{
		"version": etc.Version,
		"user":    user,
		"base":    idata.Config.HTTPBase,
		"name":    oauth2.ProviderIDMap[user.Provider].NamePrefix + user.Name,
	})
}

func HandleSearchAPI(w http.ResponseWriter, r *http.Request) {
	_, user, err := ApiBootstrap(r, w, []string{http.MethodGet}, config.GlobalSearchOff, config.GlobalSearchOff, false)
	if err != nil {
		WriteJSON(w, map[string]interface{}{
			"response": "bad",
			"message":  err.Error(),
		})
		return
	}
	q := db.DB.Build().Se("*").Fr("files")
	{
		qq := r.Form.Get("q")
		wh := r.Form.Get("w")
		if len(qq) > 0 {
			lol := q.WR("path", "like", "'%'||?||'%'", true, qq)
			if len(wh) > 0 {
				lol = lol.WR("path", "like", "'%'||?||'%'", true, wh)
			}
		}
	}
	for _, item := range []string{"md5", "sha1", "sha256", "sha512", "sha3", "blake2b"} {
		qh := r.Form.Get(item)
		if len(qh) > 0 {
			q.Wh("hash_"+item, qh)
		}
	}
	fa1 := db.File{}.ScanAll(q.Lm(25))
	ua := user.GetAccess()
	fa2 := []*db.File{}
	//
	for _, item := range fa1 {
		if _, ok := idata.DataPathsPub[item.Root]; ok {
			fa2 = append(fa2, item)
			continue
		}
		if _, ok := idata.DataPathsPrv[item.Root]; ok {
			for _, jtem := range ua {
				if strings.HasPrefix(item.Path, jtem) {
					fa2 = append(fa2, item)
					continue
				}
			}
		}
	}
	WriteJSON(w, map[string]interface{}{
		"response": "good",
		"count":    len(fa2),
		"results":  fa2,
	})
}
