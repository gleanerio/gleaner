package acquire

import (
	"github.com/samclarke/robotstxt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func getRobotsTxt(robotsUrl string) (*robotstxt.RobotsTxt, error) {
	var client http.Client

	req, err := http.NewRequest("GET", robotsUrl, nil)
	if err != nil {
		log.Printf("error creating http request: %s  ", err)
		return nil, err
	}
	req.Header.Set("User-Agent", EarthCubeAgent)
	req.Header.Set("Accept", "text/plain, text/html")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error fetching robots.txt at %s : %s  ", robotsUrl, err)
		return nil, err
	}

	if resp.StatusCode >= 400 {
		log.Printf("Robots.txt unavailable at %s", robotsUrl)
		return nil, nil
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response for robots.txt at %s : %s ", robotsUrl, err)
		return nil, err
	}

	robots, err := robotstxt.Parse(string(bodyBytes), robotsUrl)
	if err != nil {
		log.Printf("error parsing robots.txt at %s : %s  ", robotsUrl, err)
		return nil, err
	}

	return robots, nil
}
