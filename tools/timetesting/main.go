package main

import (
	"fmt"
	"sort"
	"time"
)

type reviews_data struct {
	review_id string
	date      time.Time
}

func main() {

	currentTime := time.Now()
	fmt.Println("Short Hour Minute Second: ", currentTime.Format("2006-01-02-3-4"))

	layout := "2006-01-02-3-4"
	str := "2014-11-12-11-12"
	t, err := time.Parse(layout, str)

	if err != nil {
		fmt.Println(err)

	}
	fmt.Println(t)

	fmt.Println("Sort Example")
	var listOfReviews = make([]reviews_data, 0)
	listOfReviews = append(listOfReviews, reviews_data{review_id: "1", date: time.Now().AddDate(0, 0, 8*1)})
	listOfReviews = append(listOfReviews, reviews_data{review_id: "2", date: time.Now().AddDate(0, 0, 9*1)})
	listOfReviews = append(listOfReviews, reviews_data{review_id: "1", date: time.Now().AddDate(0, 0, 7*-1)})

	sort.Slice(listOfReviews, func(i, j int) bool { return listOfReviews[i].date.Before(listOfReviews[j].date) })

	fmt.Println(listOfReviews)

}
