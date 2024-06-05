package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/samicrusader/andesite-the-eye.eu/pkg/db"
	"github.com/samicrusader/andesite-the-eye.eu/pkg/idata"

	"github.com/gorilla/sessions"
	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/util"
	discord "github.com/nektro/go.discord"
	etc "github.com/nektro/go.etc"

	. "github.com/nektro/go-util/alias"
)

func Filter(stack []os.FileInfo, cb func(os.FileInfo) bool) []os.FileInfo {
	result := []os.FileInfo{}
	for _, item := range stack {
		if cb(item) {
			result = append(result, item)
		}
	}
	return result
}

func WriteUserDenied(r *http.Request, w http.ResponseWriter, fileOrAdmin bool, showLogin bool) {
	me := ""
	sess := etc.GetSession(r)
	sessName := sess.Values["name"]
	if sessName != nil {
		sessID := sess.Values["user"].(string)
		provider := sess.Values["provider"].(string)
		me += F(" %s@%s (%s)", sessName.(string), provider, sessID)
	}

	message := ""
	if fileOrAdmin {
		if showLogin {
			message = "You " + me + " do not have access to this resource."
		} else {
			message = "Unable to find the requested resource for you" + me + "."
		}
	} else {
		message = "Admin priviledge required. Access denied."
	}

	linkmsg := ""
	if showLogin {
		linkmsg = "Please <a href='" + idata.Config.HTTPBase + "login'>Log In</a>."
		w.WriteHeader(http.StatusForbidden)
		WriteResponse(r, w, "Forbidden", message, linkmsg)
	} else {
		linkmsg = "<a href='" + idata.Config.HTTPBase + "logout'>Logout</a>."
		w.WriteHeader(http.StatusForbidden)
		WriteResponse(r, w, "Not Found", message, linkmsg)
	}
}

func WriteAPIResponse(r *http.Request, w http.ResponseWriter, good bool, message string) {
	if !good {
		w.WriteHeader(http.StatusForbidden)
	}
	titlemsg := ""
	if good {
		titlemsg = "Update Successful"
	} else {
		titlemsg = "Update Failed"
	}
	WriteResponse(r, w, titlemsg, message, "Return to <a href='"+idata.Config.HTTPBase+"admin'>the dashboard</a>.")
}

func WriteResponse(r *http.Request, w http.ResponseWriter, title string, message string, link string) {
	etc.WriteHandlebarsFile(r, w, "/response.hbs", map[string]interface{}{
		"version": etc.Version,
		"title":   title,
		"message": message,
		"link":    link,
		"base":    idata.Config.HTTPBase,
	})
}

func WriteLinkResponse(r *http.Request, w http.ResponseWriter, title string, message string, linkText string, href string) {
	WriteResponse(r, w, title, message, "<a href=\""+href+"\">"+linkText+"</a>")
}

func ContainsAll(mp url.Values, keys ...string) bool {
	for _, item := range keys {
		if _, ok := mp[item]; !ok {
			return false
		}
	}
	return true
}

func ApiBootstrap(r *http.Request, w http.ResponseWriter, methods []string, requireLogin bool, requireAdmin bool, doOutput bool) (*sessions.Session, *db.User, error) {
	if !stringsu.Contains(methods, r.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Add("Allow", F("%v", methods))
		if doOutput {
			WriteAPIResponse(r, w, false, "This action requires using HTTP "+F("%v", methods))
		}
		return nil, nil, E("")
	}

	sess := etc.GetSession(r)
	provID := sess.Values["provider"]
	sessID := sess.Values["user"]

	if requireLogin && sessID == nil {
		pk := ""

		if len(pk) == 0 {
			pk = r.Header.Get("x-passkey")
		}
		if len(pk) == 0 {
			u, _, o := r.BasicAuth()
			if o && len(u) > 0 {
				pk = u
			}
		}
		if len(pk) == 0 {
			if doOutput {
				WriteUserDenied(r, w, true, true)
			}
			return nil, nil, E("not logged in and no passkey found")
		}
		u, ok := db.User{}.ByPasskey(pk)
		if !ok {
			if doOutput {
				WriteUserDenied(r, w, true, true)
			}
			return nil, nil, E("invalid passkey")
		}
		provID = u.Provider
		sessID = u.Snowflake
	}
	var pS, uS string
	if provID != nil {
		pS = provID.(string)
	}
	if sessID != nil {
		uS = sessID.(string)
	}
	user, ok := db.User{}.BySnowflake(pS, uS)

	if requireLogin {
		if !ok {
			if doOutput {
				WriteResponse(r, w, "Access Denied", "This action requires being a member of this server. ("+uS+"@"+pS+")", "")
			}
			return nil, nil, E("")
		}
		if requireAdmin && !user.Admin {
			if doOutput {
				WriteAPIResponse(r, w, false, "This action requires being a site administrator. ("+uS+"@"+pS+")")
			}
			return nil, nil, E("")
		}
	} else {
		if !ok {
			user = &db.User{ID: -1, Name: "Guest", Provider: r.Host}
		}
	}

	err := r.ParseForm()
	if err != nil {
		if doOutput {
			WriteAPIResponse(r, w, false, "Error parsing form data")
		}
		return nil, nil, E("")
	}

	return sess, user, nil
}

func WriteJSON(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("content-type", "application/json")
	bytes, _ := json.Marshal(data)
	fmt.Fprintln(w, string(bytes))
}

func MakeDiscordRequest(endpoint string, body url.Values) []byte {
	req, _ := http.NewRequest(http.MethodGet, idata.DiscordAPI+endpoint, strings.NewReader(body.Encode()))
	req.Header.Set("User-Agent", "nektro/andesite")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bot "+idata.Config.GetDiscordClient().Extra2)
	req.Header.Set("Accept", "application/json")
	return util.DoHttpRequest(req)
}

func FetchDiscordRole(guild string, role string) discord.GuildRole {
	bys := MakeDiscordRequest("/guilds/"+guild+"/roles", url.Values{})
	roles := []discord.GuildRole{}
	json.Unmarshal(bys, &roles)
	for i, item := range roles {
		if item.ID == role {
			return roles[i]
		}
	}
	return discord.GuildRole{}
}

type DiscordGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func FetchDiscordGuild(guild string) DiscordGuild {
	bys := MakeDiscordRequest("/guilds/"+guild, url.Values{})
	var dg DiscordGuild
	json.Unmarshal(bys, &dg)
	return dg
}

func MapToArray(mp map[string]string) [][]string {
	result := [][]string{}
	for k, v := range mp {
		result = append(result, []string{k, v})
	}
	return result
}

func Combine(mps ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, item := range mps {
		for k, v := range item {
			result[k] = v
		}
	}
	return result
}
