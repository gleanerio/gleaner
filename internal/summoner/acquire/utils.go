package acquire

import (
	"log"
    "net/http"
    "io/ioutil"
    "github.com/samclarke/robotstxt"
)

func getRobotsTxt(robotsUrl string) (robotstxt, error) {
	var client http.Client
	var robots robotstxt.RobotsTxt
    req, err := http.NewRequest("GET", robotsUrl, nil)
    if err != nil {
        log.Printf("error creating http request: %s  ",  err)
        return robots, err
    }
    req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")
    req.Header.Set("Accept", "text/plain, text/html")

    resp, err := client.Do(req)
    if err != nil {
        log.Printf("error fetching robots.txt at %s : %s  ", robotsUrl, err)
        return robots, err
    }
    defer resp.Body.Close()
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("error reading response for robots.txt at %s : %s ", robotsUrl, err)
        return robots, err
    }

    robots, err = robotstxt.Parse(string(bodyBytes), robotsUrl)
    if err != nil {
        log.Printf("error parsing robots.txt at %s : %s  ", robotsUrl, err)
        return robots, err
    }

    return robots, nil
}
