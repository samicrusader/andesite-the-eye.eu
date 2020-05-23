package main

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/nektro/andesite/pkg/db"
	"github.com/nektro/andesite/pkg/fsdb"
	"github.com/nektro/andesite/pkg/handler"
	"github.com/nektro/andesite/pkg/idata"

	"github.com/aymerick/raymond"
	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/util"
	etc "github.com/nektro/go.etc"
	"github.com/nektro/go.etc/htp"
	"github.com/spf13/pflag"

	. "github.com/nektro/go-util/alias"

	_ "github.com/nektro/andesite/statik"
)

var (
	Version = "vMASTER"
)

func main() {
	idata.Version = etc.FixBareVersion(Version)
	util.Log("Initializing Andesite " + idata.Version + "...")
	etc.AppID = "andesite"

	pflag.IntVar(&idata.Config.Version, "version", idata.RequiredConfigVersion, "Config version to use.")
	pflag.StringVar(&idata.Config.Root, "root", "", "Path of root directory for files")
	pflag.IntVar(&idata.Config.Port, "port", 8000, "Port to open server on")
	pflag.StringVar(&idata.Config.HTTPBase, "base", "/", "Http Origin Path")
	pflag.StringVar(&idata.Config.Public, "public", "", "Public root of files to serve")
	pflag.StringArrayVar(&idata.Config.SearchOn, "enable-search", []string{}, "Set to a root ID to enable file search for that directory.")
	pflag.StringArrayVar(&idata.Config.SearchOff, "disable-search", []string{}, "Set to a root ID to disable file search for that directory.")
	flagDGS := pflag.String("discord-guild-id", "", "")
	flagDBT := pflag.String("discord-bot-token", "", "")
	pflag.BoolVar(&idata.Config.Verbose, "verbose", false, "")
	pflag.BoolVar(&idata.Config.VerboseFS, "fsdb-verbose", false, "")
	etc.PreInit()

	etc.Init("andesite", &idata.Config, "./files/", db.SaveOAuth2InfoCb)

	//

	for i, item := range idata.Config.Clients {
		if item.For == "discord" {
			if len(*flagDGS) > 0 {
				idata.Config.Clients[i].Extra1 = *flagDGS
			}
			if len(*flagDBT) > 0 {
				idata.Config.Clients[i].Extra2 = *flagDBT
			}
		}
	}

	if idata.Config.Version == 0 {
		idata.Config.Version = 1
	}
	if idata.Config.Version != idata.RequiredConfigVersion {
		util.DieOnError(
			E(F("Current idata.Config.json version '%d' does not match required version '%d'.", idata.Config.Version, idata.RequiredConfigVersion)),
			F("Visit https://github.com/nektro/andesite/blob/master/docs/config/v%d.md for more info.", idata.RequiredConfigVersion),
		)
	}

	idata.Config.SearchOn = stringsu.Depupe(idata.Config.SearchOn)
	idata.Config.SearchOff = stringsu.Depupe(idata.Config.SearchOff)

	//
	// database initialization

	db.Init()

	db.Upgrade()

	//
	// graceful stop

	util.RunOnClose(func() {
		util.Log("Gracefully shutting down...")

		util.Log("Saving database to disk")
		db.DB.Close()

		util.Log("Done!")
	})

	//
	// handlebars helpers

	raymond.RegisterHelper("url_name", func(x string) string {
		return strings.Replace(url.PathEscape(x), "%2F", "/", -1)
	})
	raymond.RegisterHelper("add_i", func(a, b int) int {
		return a + b
	})

	//
	// http server setup

	htp.Register("/test", "GET", handler.HandleTest)

	if len(idata.Config.Root) > 0 {
		idata.Config.Root, _ = filepath.Abs(filepath.Clean(strings.ReplaceAll(idata.Config.Root, "~", idata.HomedirPath)))
		util.DieOnError(util.Assert(util.DoesDirectoryExist(idata.Config.Root), "Please pass a valid directory as a root parameter!"))
		idata.DataPathsPrv["files"] = idata.Config.Root
	}
	for _, item := range idata.Config.RootsPrv {
		ab, err := filepath.Abs(item[1])
		util.DieOnError(err)
		idata.DataPathsPrv[item[0]] = ab
	}
	if len(idata.DataPathsPrv) > 0 {
		for k, v := range idata.DataPathsPrv {
			htp.Register("/"+k+"/*", "GET", handler.HandleDirectoryListing(handler.HandleFileListing))
			util.Log("Sharing private files as", k, "from ", v)
		}

		htp.Register("/regen_passkey", "GET", handler.HandleRegenPasskey)
		htp.Register("/logout", "GET", handler.HandleLogout)
		htp.Register("/open/*", "GET", handler.HandleDirectoryListing(handler.HandleShareListing))

		htp.Register("/admin", "GET", handler.HandleAdmin)
		htp.Register("/admin/users", "GET", handler.HandleAdminUsers)
		htp.Register("/admin/roots", "GET", handler.HandleAdminRoots)

		htp.Register("/api/access/create", "POST", handler.HandleAccessCreate)
		htp.Register("/api/access/update", "POST", handler.HandleAccessUpdate)
		htp.Register("/api/access/delete", "POST", handler.HandleAccessDelete)

		htp.Register("/api/share/create", "POST", handler.HandleShareCreate)
		htp.Register("/api/share/update", "POST", handler.HandleShareUpdate)
		htp.Register("/api/share/delete", "POST", handler.HandleShareDelete)

		htp.Register("/api/access_discord_role/create", "POST", handler.HandleDiscordRoleAccessCreate)
		htp.Register("/api/access_discord_role/update", "POST", handler.HandleDiscordRoleAccessUpdate)
		htp.Register("/api/access_discord_role/delete", "POST", handler.HandleDiscordRoleAccessDelete)
	}

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
	if len(idata.DataPathsPub) > 0 {
		for k, v := range idata.DataPathsPub {
			htp.Register("/"+k+"/*", "GET", handler.HandleDirectoryListing(handler.HandlePublicListing))
			util.Log("Sharing public files as", k, "from ", v)
		}
	}

	//
	// initialize file database in background

	htp.Register("/search", "GET", handler.HandleSearch)
	htp.Register("/api/search", "GET", handler.HandleSearchAPI)

	if len(idata.Config.SearchOn) > 0 {
		for _, item := range idata.Config.SearchOn {
			go fsdb.Init(idata.DataPathsPub, item)
			go fsdb.Init(idata.DataPathsPrv, item)
		}
	}
	if len(idata.Config.SearchOff) > 0 {
		for _, item := range idata.Config.SearchOff {
			fsdb.DeInit(idata.DataPathsPub, item)
			fsdb.DeInit(idata.DataPathsPrv, item)
		}
	}

	//
	// start http server

	etc.StartServer(idata.Config.Port)
}
