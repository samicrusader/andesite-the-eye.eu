package handler

import (
	"net/http"

	"github.com/samicrusader/andesite-the-eye.eu/pkg/db"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"

	etc "github.com/nektro/go.etc"
	oauth2 "github.com/nektro/go.oauth2"
)

// handler for http://andesite/admin
func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	_, user, err := ApiBootstrap(r, w, []string{http.MethodGet}, true, true, true)
	if err != nil {
		return
	}
	dc := idata.Config.GetDiscordClient()
	etc.WriteHandlebarsFile(r, w, "/admin.hbs", map[string]interface{}{
		"version":               etc.Version,
		"user":                  user,
		"base":                  idata.Config.HTTPBase,
		"name":                  oauth2.ProviderIDMap[user.Provider].NamePrefix + user.Name,
		"auth":                  oauth2.ProviderIDMap[user.Provider].ID,
		"discord_role_share_on": len(dc.Extra1) > 0 && len(dc.Extra2) > 0,
		"users":                 db.User{}.All(),
		"accesses":              db.UserAccess{}.All(),
		"shares":                db.Share{}.All(),
		"discord_shares":        db.DiscordRoleAccess{}.All(),
	})
}

// handler for http://andesite/admin/users
func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	_, user, err := ApiBootstrap(r, w, []string{http.MethodGet}, true, true, true)
	if err != nil {
		return
	}
	etc.WriteHandlebarsFile(r, w, "/users.hbs", map[string]interface{}{
		"version": etc.Version,
		"user":    user,
		"base":    idata.Config.HTTPBase,
		"name":    oauth2.ProviderIDMap[user.Provider].NamePrefix + user.Name,
		"auth":    oauth2.ProviderIDMap[user.Provider].ID,
		"users":   db.User{}.All(),
	})
}

func HandleAdminRoots(w http.ResponseWriter, r *http.Request) {
	_, user, err := ApiBootstrap(r, w, []string{http.MethodGet}, true, true, true)
	if err != nil {
		return
	}
	etc.WriteHandlebarsFile(r, w, "/admin_roots.hbs", map[string]interface{}{
		"version":       etc.Version,
		"user":          user,
		"base":          idata.Config.HTTPBase,
		"roots_public":  MapToArray(idata.DataPathsPub),
		"roots_private": MapToArray(idata.DataPathsPrv),
	})
}
