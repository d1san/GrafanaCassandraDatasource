package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"
)

type QueryProcessor struct{}

func (qp *QueryProcessor) processRawMetricQuery(result *datasource.QueryResult, query string, ds *CassandraDatasource) {

	iter := ds.session.Query(query).Iter()

	var id string
	var timestamp time.Time
	var value float64

	var seriesName string
	if len(iter.Columns()) > 2 {
		valueColumnName := iter.Columns()[1].Name
		if !strings.HasPrefix(valueColumnName, "cast") {
			seriesName = valueColumnName
		}
	}

	series := make(map[string]*datasource.TimeSeries)

	for iter.Scan(&id, &value, &timestamp) {
		if _, ok := series[id]; !ok {
			name := id
			if seriesName != "" {
				name = seriesName
			}

			series[id] = &datasource.TimeSeries{Name: name}
		}

		series[id].Points = append(series[id].Points, &datasource.Point{
			Timestamp: timestamp.UnixNano() / int64(time.Millisecond),
			Value:     value,
		})

	}

	if err := iter.Close(); err != nil {
		ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
		result.Error = err.Error()

		return
	}

	for _, serie2 := range series {
		result.Series = append(result.Series, serie2)
	}
}

func (qp *QueryProcessor) processStrictMetricQuery(result *datasource.QueryResult, query string, valueId string, ds *CassandraDatasource) {

	iter := ds.session.Query(query).Iter()

	var timestamp time.Time
	var value float64

	serie := &datasource.TimeSeries{Name: valueId}

	for iter.Scan(&timestamp, &value) {
		serie.Points = append(serie.Points, &datasource.Point{
			Timestamp: timestamp.UnixNano() / int64(time.Millisecond),
			Value:     value,
		})
	}
	if err := iter.Close(); err != nil {
		ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
		result.Error = err.Error()

		return
	}

	result.Series = append(result.Series, serie)
}
