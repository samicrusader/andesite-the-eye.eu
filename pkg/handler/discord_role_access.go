package handler

import (
	"net/http"

	"github.com/samicrusader/andesite-the-eye.eu/pkg/db"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"

	"github.com/nektro/go.etc/htp"

	. "github.com/nektro/go-util/alias"
)

func HandleDiscordRoleAccessCreate(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	_, _, err := ApiBootstrap(r, w, []string{http.MethodPost}, true, true, true)
	if err != nil {
		return
	}
	// ags := c.GetFormString("GuildID")
	ags := idata.Config.GetDiscordClient().Extra1
	agr := c.GetFormString("RoleID")
	apt := c.GetFormString("Path")
	gn := FetchDiscordGuild(ags).Name
	rn := FetchDiscordRole(ags, agr).Name
	if len(gn) == 0 && len(rn) == 0 {
		WriteAPIResponse(r, w, false, "Unable to fetch role metadata from Discord API.")
		return
	}
	db.CreateDiscordRoleAccess(ags, agr, apt, gn, rn)
	WriteAPIResponse(r, w, true, F("Created access for %s / %s to %s.", gn, rn, apt))
}

func HandleDiscordRoleAccessUpdate(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	_, _, err := ApiBootstrap(r, w, []string{http.MethodPost}, true, true, true)
	if err != nil {
		return
	}
	if !ContainsAll(r.PostForm, "ID", "RoleID", "Path") {
		WriteAPIResponse(r, w, false, "Missing POST values")
		return
	}
	_, qid := c.GetFormInt("id")
	// qgs := c.GetFormString("GuildID")
	qgs := idata.Config.GetDiscordClient().Extra1
	qgr := c.GetFormString("RoleID")
	qpt := c.GetFormString("Path")
	gn := FetchDiscordGuild(qgs).Name
	rn := FetchDiscordRole(qgs, qgr).Name
	if len(gn) == 0 && len(rn) == 0 {
		WriteAPIResponse(r, w, false, "Unable to fetch role metadata from Discord API.")
		return
	}
	dra, ok := db.DiscordRoleAccess{}.ByID(qid)
	c.Assert(ok, "400: unable to fine the DiscordRoleAccess with that ID")
	dra.SetGuildID(qgs)
	dra.SetGuildName(gn)
	dra.SetRoleID(qgr)
	dra.SetRoleName(rn)
	dra.SetPath(qpt)
	WriteAPIResponse(r, w, true, F("Successfully updated share path for %s / %s to %s.", gn, rn, qpt))
}

func HandleDiscordRoleAccessDelete(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	_, _, err := ApiBootstrap(r, w, []string{http.MethodPost}, true, true, true)
	if err != nil {
		return
	}
	_, qID := c.GetFormInt("id")
	dra, ok := db.DiscordRoleAccess{}.ByID(qID)
	c.Assert(ok, "400: unable to fine the DiscordRoleAccess with that ID")
	dra.Delete()
	WriteAPIResponse(r, w, true, F("Successfully deleted access for %s / %s to %s.", dra.GuildName, dra.RoleName, dra.Path))
}
