// +build ignore

package main

import (
	"fmt"
	"github.com/wcharczuk/go-chart"
	"github.com/willf/bloom"
	"net/http"
)

// Used to find optimal standard m and k values

func getSeriesValues(m uint, k uint, n uint) chart.ContinuousSeries {
	entries := 128
	xValues := make([]float64, entries, entries)
	yValues := make([]float64, entries, entries)
	for i := 0; i < entries; i++ {
		// m := quasar.StandardConfig.FiltersM
		// k := quasar.StandardConfig.FiltersK

		n := i * 16
		xValues[i] = float64(n)
		yValues[i] = bloom.New(m, k).EstimateFalsePositiveRate(uint(n))
	}

	return chart.ContinuousSeries{
		Name:    fmt.Sprintf("m=%d, k=%d, n=%d", m, k, n),
		XValues: xValues,
		YValues: yValues,
	}
}

func drawChart(res http.ResponseWriter, req *http.Request) {

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.Style{Show: true},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{Show: true},
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 260,
			},
		},
		Series: []chart.Series{
			getSeriesValues(1024*8, 6, 1024),
			getSeriesValues(1024*8, 3, 2048),
			getSeriesValues(1280*8, 7, 1024),
			getSeriesValues(1280*8, 3, 2048),
		},
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func main() {
	http.HandleFunc("/", drawChart)
	http.ListenAndServe(":8080", nil)
}
