package acquire

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
	"github.com/mafredri/cdp/session"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var HEADLESS_URL = "http://127.0.0.1:9222"

func PingHeadless() (int, error) {
	var client = http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("HEAD", HEADLESS_URL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func TestHeadlessNG(t *testing.T) {
	status, err := PingHeadless()

	if err != nil || status != 200 {
		t.Skip("Skipping headless tests because no headless browser is running.")
	}

	tests := []struct {
		name         string
		url          string
		jsonldcount  int
		headlessWait int
		expectedFail bool "default:false"
	}{
		{name: "r2r_wait_5_works_returns_2_jsonld",
			url:          "https://dev.rvdata.us/search/fileset/100135",
			jsonldcount:  2,
			headlessWait: 20,
		},
		{name: "r2r_expectedfail_wait_0_returns_1_jsonld_fails_if_2_jsonld",
			url:          "https://dev.rvdata.us/search/fileset/100135",
			jsonldcount:  2,
			headlessWait: 0,
			expectedFail: true,
		},
	}

	for _, test := range tests {

		runstats := common.NewRepoStats(test.name)
		conf := map[string]interface{}{
			"minio":    map[string]interface{}{"bucket": "test"},
			"summoner": map[string]interface{}{"threads": "5", "delay": 10, "headless": HEADLESS_URL},
			"sources":  []map[string]interface{}{{"name": test.name, "headlessWait": test.headlessWait}},
		}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		ctx, cancel := chromedp.NewContext(context.TODO())
		defer cancel()

		// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
		//devt := devtool.New(mcfg["headless"])
		devt := devtool.New(HEADLESS_URL)

		pt, err := devt.Get(ctx, devtool.Page)
		if err != nil {
			pt, err = devt.Create(ctx)
			if err != nil {
				log.WithFields(log.Fields{"issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)
				return
			}
		}

		// Initiate a new RPC connection to the Chrome DevTools Protocol target.
		conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
		if err != nil {
			log.WithFields(log.Fields{"issue": "Not REPO FAULT. Devtools... Is Headless Container running?"}).Error(err)

			return
		}
		defer conn.Close()
		sessionclient := cdp.NewClient(conn)
		manager, err := session.NewManager(sessionclient)
		if err != nil {
			// Handle error.
		}
		defer manager.Close()
		args := target.NewCreateTargetArgs("")
		//args.SetNewWindow(true)
		newPage, err := sessionclient.Target.CreateTarget(ctx,
			args)
		if err != nil {
			log.WithFields(log.Fields{"url": test.url, "issue": "Not REPO FAULT. NewCreateTargetArgs... Is Headless Container running?"}).Error(err)

			return
		}
		closeArgs := target.NewCloseTargetArgs(newPage.TargetID)
		defer func(Target cdp.Target, ctx context.Context, args *target.CloseTargetArgs) {
			log.Info("Close Target Defer")
			_, err := Target.CloseTarget(ctx, args)
			if err != nil {
				log.WithFields(log.Fields{"url": test.url, "issue": "error closing target"}).Error("PageRenderAndUpload ", test.url, " ::", err)

			}
		}(sessionclient.Target, ctx, closeArgs)
		repoLogger, _ := common.LogIssues(viper, test.name)
		t.Run(test.name, func(t *testing.T) {
			jsonlds, err := PageRender(viper, 60*time.Second, test.url, test.name, repoLogger, runstats, manager, newPage.TargetID)
			if !test.expectedFail {
				assert.Equal(t, test.jsonldcount, len(jsonlds))
			} else {
				assert.NotEqual(t, test.jsonldcount, len(jsonlds))
			}

			assert.Nil(t, err)

		})
	}
}
