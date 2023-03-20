package acquire

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/temoto/robotstxt"
	"net/http"
)

func getRobotsTxt(robotsUrl string) (*robotstxt.RobotsData, error) {
	var client http.Client

	req, err := http.NewRequest("GET", robotsUrl, nil)
	if err != nil {
		log.Error("error creating http request: ", err)
		return nil, err
	}
	req.Header.Set("User-Agent", EarthCubeAgent)
	req.Header.Set("Accept", "text/plain, text/html")

	resp, err := client.Do(req)
	if err != nil {
		log.Info("error fetching robots.txt at ", robotsUrl, err)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Robots.txt unavailable at %s", robotsUrl)
	}

	defer resp.Body.Close()

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		log.Error("error parsing robots.txt at ", robotsUrl, err)
		return nil, err
	}
	return robots, nil
}
