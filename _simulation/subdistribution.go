// +build ignore

package main

import (
	"github.com/wcharczuk/go-chart"
	"math"
	"math/rand"
	"net/http"
	"sort"
)

func getTopic() int {
	// vaguely based on twitter distribution
	x := rand.NormFloat64() * rand.NormFloat64() * rand.NormFloat64()
	return int(math.Abs(x * 10000.0))
}

func getDistribution(totalSubs int, topicLimit int) ([]float64, []float64) {
	subCount := make(map[int]int)
	for i := 0; i < totalSubs; i++ {
		topic := getTopic()
		if count, ok := subCount[topic]; ok {
			subCount[topic] = count + 1
		} else {
			subCount[topic] = 1
		}
	}

	topics := []int{}
	for topic, _ := range subCount {
		if topicLimit != 0 && topic >= topicLimit {
			continue
		}
		topics = append(topics, topic)
	}
	sort.Ints(topics)

	xValues := make([]float64, len(topics), len(topics))
	yValues := make([]float64, len(topics), len(topics))
	for i, topic := range topics {
		xValues[i] = float64(topic)
		yValues[i] = float64(subCount[topic])
	}

	return xValues, yValues
}

func drawChart(res http.ResponseWriter, req *http.Request) {

	// vaguely based on twitter
	users := 328000000
	avgSubs := 108
	total := users * avgSubs
	xValues, yValues := getDistribution(total/1000, 100)

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:      "Topics",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "Subscriptions",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func main() {
	http.HandleFunc("/", drawChart)
	http.ListenAndServe(":8080", nil)
}
