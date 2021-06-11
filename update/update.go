package update

import (
	"net/url"
	"path"
	"time"

	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

func Check(cfg *config.Config, force bool) bool {
	entry, ok := cfg.GetDefaultEntry()
	if !ok {
		tools.Log.Error().Msg("unable to get default configuration entry")
		return false
	}

	if !force {
		diff := time.Since(entry.LastUpdateCheck)
		if diff < 8*time.Hour {
			// only checking periodically to avoid nagging (and avoid thundering herd updates)
			return false
		}
	}

	updater := getUpdater(entry, app.Version)

	newVersion, err := updater.UpdateAvailable()
	if err != nil {
		tools.Log.Err(err).Msg("unable to check for updates")
		return false
	}

	entry.LastUpdateCheck = time.Now()
	config.SetEntry(entry)

	if newVersion != "" {
		tools.LogStderr(nil,
			"\n*** Notice: A new version is available (%s to %s). "+
				"Use update to get the latest version. ***\n\n",
			app.Version, newVersion)
		return true
	}

	return false
}

func GetLatest(cfg *config.Config) error {
	entry, _ := cfg.GetDefaultEntry()
	updater := getUpdater(entry, app.Version)

	return updater.BackgroundRun()
}

func getUpdater(entry *config.Entry, currentVersion string) *selfupdate.Updater {
	updateURL, _ := url.Parse(entry.ServiceUrl)
	updateURL.Path = "/api/v1/apps/cli/"
	return &selfupdate.Updater{
		CurrentVersion: currentVersion,
		ApiURL:         updateURL.String(),
		BinURL:         updateURL.String(),
		DiffURL:        updateURL.String(),
		Dir:            path.Dir(config.DefaultConfigFile),
		CmdName:        "aardy",
	}
}
