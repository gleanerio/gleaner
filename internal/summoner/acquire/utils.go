package acquire

import (
	"fmt"
	"github.com/samclarke/robotstxt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func getRobotsTxt(robotsUrl string) (*robotstxt.RobotsTxt, error) {
	var client http.Client

	req, err := http.NewRequest("GET", robotsUrl, nil)
	if err != nil {
		log.Error("error creating http request:", err)
		return nil, err
	}
	req.Header.Set("User-Agent", EarthCubeAgent)
	req.Header.Set("Accept", "text/plain, text/html")

	resp, err := client.Do(req)
	if err != nil {
		log.Error("error fetching robots.txt at", robotsUrl, err)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Robots.txt unavailable at %s", robotsUrl)
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error reading response for robots.txt at", robotsUrl, err)
		return nil, err
	}

	robots, err := robotstxt.Parse(string(bodyBytes), robotsUrl)
	if err != nil {
		log.Error("error parsing robots.txt at", robotsUrl, err)
		return nil, err
	}

	return robots, nil
}
