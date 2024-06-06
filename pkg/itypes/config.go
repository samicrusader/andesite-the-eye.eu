package itypes

import (
	oauth2 "github.com/nektro/go.oauth2"
)

type Config struct {
	Version   int               `json:"version"`
	Root      string            `json:"root"`
	Public    string            `json:"public"`
	Themes    []string          `json:"themes"`
	HTTPBase  string            `json:"base"`
	Clients   []oauth2.AppConf  `json:"clients"`
	Providers []oauth2.Provider `json:"providers"`
	SearchOn  []string          `json:"search_on"`
	SearchOff []string          `json:"search_off"`
	Verbose   bool              `json:"verbose"`
	VerboseFS bool              `json:"verbose_fsdb"`
	RootsPub  [][]string        `json:"roots_public"`
	RootsPrv  [][]string        `json:"roots_private"`
	OffHashes []string
	ScanSimul int
	CRootsPub []string
	CRootsPrv []string
}

func (c *Config) GetDiscordClient() *oauth2.AppConf {
	for _, item := range c.Clients {
		if item.For == "discord" {
			return &item
		}
	}
	return &oauth2.AppConf{}
}
