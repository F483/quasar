// +build ignore

package main

import (
	"github.com/f483/quasar"
	"github.com/wcharczuk/go-chart"
	"github.com/willf/bloom"
	"net/http"
)

func getChartData() ([]float64, []float64) {
	entries := 128
	xValues := make([]float64, entries, entries)
	yValues := make([]float64, entries, entries)
	for i := 0; i < entries; i++ {
		n := i * 16
		m := quasar.StandardConfig.FiltersM
		k := quasar.StandardConfig.FiltersK
		f := bloom.New(uint(m), uint(k))
		fpRate := f.EstimateFalsePositiveRate(uint(n))
		xValues[i] = float64(n)
		yValues[i] = fpRate
	}
	return xValues, yValues
}

func drawChart(res http.ResponseWriter, req *http.Request) {

	xValues, yValues := getChartData()

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:      "Node topic subscriptions.",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Name:      "Node false positive rate.",
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
