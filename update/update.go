package update

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/sanbornm/go-selfupdate/selfupdate"

	"github.com/1xyz/pryrite/app"
	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/tools"
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

	// always update the check date, even if we might fail below, to avoid constant checks
	// when offline, etc. we'll just try again in X hours, no rush
	entry.LastUpdateCheck = time.Now()
	config.SetEntry(entry)

	updater := getUpdater(entry, app.Version)

	newVersion, err := updater.UpdateAvailable()
	if err != nil {
		tools.Log.Err(err).Msg("unable to check for updates")
		return false
	}

	if newVersion != "" {
		tools.LogStderr(nil,
			"\n*** Notice: A new version is available (%s to %s). "+
				"Use update to get the latest version. ***\n\n",
			app.Version, newVersion)
		return true
	}

	return false
}

func GetLatest(cfg *config.Config) (string, error) {
	entry, _ := cfg.GetDefaultEntry()
	updater := getUpdater(entry, app.Version)

	err := updater.BackgroundRun()
	if err != nil {
		return "", err
	}

	switch updater.Result {
	case selfupdate.PatchResult:
		return updater.Info.Version + " (patched)", nil
	case selfupdate.FullBinResult:
		return updater.Info.Version, nil
	case selfupdate.AtLatestResult:
		return "", errors.New(updater.Info.Version + " already installed")
	default:
		return "", errors.New("failed to install new version: See log file for details")
	}
}

func getUpdater(entry *config.Entry, currentVersion string) *selfupdate.Updater {
	updateURL, _ := url.Parse(entry.ServiceUrl)
	updateURL.Path = "/api/v1/apps/cli/"
	return &selfupdate.Updater{
		CurrentVersion: currentVersion,
		ApiURL:         updateURL.String(),
		BinURL:         updateURL.String(),
		DiffURL:        updateURL.String(),
		Dir:            filepath.Dir(config.DefaultConfigFile),
		CmdName:        "aardy",
		Requester:      &requestor{},
	}
}

//--------------------------------------------------------------------------------
// provide a timeout for HTTP requests...

type requestor struct{}

func (r *requestor) Fetch(url string) (io.ReadCloser, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 1 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 3 * time.Second,
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
