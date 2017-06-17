// +build ignore

package main

import (
	"github.com/f483/quasar"
	"github.com/wcharczuk/go-chart"
	"github.com/willf/bloom"
	"math"
	"math/rand"
	"net/http"
)

func getTopic() int {
	// vaguely based on twitter distribution
	x := rand.NormFloat64() * rand.NormFloat64() * rand.NormFloat64()
	return int(math.Abs(x * 10000.0))
}

func getChartData(subLimit int) ([]float64, []float64) {
	xValues := make([]float64, subLimit, subLimit)
	yValues := make([]float64, subLimit, subLimit)
	for n := 0; n < subLimit; n++ {
		m := quasar.StandardConfig.FiltersM
		k := quasar.StandardConfig.FiltersK
		f := bloom.New(uint(m), uint(k))
		fpRate := f.EstimateFalsePositiveRate(uint(n))
		xValues[n] = float64(n)
		yValues[n] = fpRate
	}
	return xValues, yValues
}

func drawChart(res http.ResponseWriter, req *http.Request) {

	xValues, yValues := getChartData(1500)

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:      "Subscriptions",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "False Positive Rate",
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
