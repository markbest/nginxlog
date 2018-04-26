package utils

import (
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
	"log"
)

type ES struct {
	esHost       string
	esIndex      string
	esType       string
	logHandle    *log.Logger
	EsClient     *elastic.Client
	boolQuery    *elastic.BoolQuery
	aggregations map[string]elastic.Aggregation
	size         int
	form         int
	sort         map[string]bool
}

// FinderResponse is the outcome of calling Finder.Find.
type ESResponse struct {
	Total int64
	Logs  []*LogFormat
	Aggs  elastic.Aggregations
}

//New ES
func NewES(host string, logHandle *log.Logger) *ES {
	//retry
	retry := elastic.NewBackoffRetrier(elastic.ZeroBackoff{})

	//connect
	client, err := elastic.NewClient(
		elastic.SetURL(host),
		elastic.SetSniff(false),
		elastic.SetTraceLog(logHandle),
		elastic.SetRetrier(retry),
	)
	if err != nil {
		panic(err)
	}
	return &ES{
		esHost:       host,
		logHandle:    logHandle,
		EsClient:     client,
		boolQuery:    elastic.NewBoolQuery(),
		aggregations: make(map[string]elastic.Aggregation),
		size:         10000,
		form:         0,
		sort:         make(map[string]bool),
	}
}

//Set ES Index
func (e *ES) Index(esIndex string) *ES {
	e.esIndex = esIndex
	return e
}

//Set ES Type
func (e *ES) Type(esType string) *ES {
	e.esType = esType
	return e
}

//ES Where Condition search
func (e *ES) Where(field string, value interface{}) *ES {
	e.boolQuery = e.boolQuery.Must(elastic.NewTermQuery(field, value))
	return e
}

//Set Size
func (e *ES) Take(size int) *ES {
	e.size = size
	return e
}

//Set Search From
func (e *ES) Page(page int) *ES {
	e.form = e.size * (page - 1)
	return e
}

//ES Range Condition search
func (e *ES) Range(field string, values map[string]interface{}) *ES {
	rangeQuery := elastic.NewRangeQuery(field)
	for k, v := range values {
		if k == "gt" {
			rangeQuery.Gt(v)
		}
		if k == "gte" {
			rangeQuery.Gte(v)
		}
		if k == "lt" {
			rangeQuery.Lt(v)
		}
		if k == "lte" {
			rangeQuery.Lte(v)
		}
	}
	e.boolQuery = e.boolQuery.Must(rangeQuery)
	return e
}

//ES Avg Aggregation
func (e *ES) Avg(field string) *ES {
	avg := elastic.NewAvgAggregation().Field(field)
	e.aggregations["avg"] = avg
	return e
}

//ES Max Aggregation
func (e *ES) Max(field string) *ES {
	max := elastic.NewMaxAggregation().Field(field)
	e.aggregations["max"] = max
	return e
}

//ES Min Aggregation
func (e *ES) Min(field string) *ES {
	min := elastic.NewMinAggregation().Field(field)
	e.aggregations["min"] = min
	return e
}

//ES Sum Aggregation
func (e *ES) Sum(field string) *ES {
	sum := elastic.NewSumAggregation().Field(field)
	e.aggregations["sum"] = sum
	return e
}

//ES Count Aggregation
func (e *ES) Count(field string) *ES {
	count := elastic.NewValueCountAggregation().Field(field)
	e.aggregations["count"] = count
	return e
}

//ES GroupBy Aggregation
func (e *ES) GroupBy(field string) *ES {
	groupBy := elastic.NewTermsAggregation().Field(field).Size(e.size)
	e.aggregations["groupBy"] = groupBy
	return e
}

//ES OrderBy Query
func (e *ES) OrderBy(field string, order string) *ES {
	var dir bool
	if order == "desc" {
		dir = false
	} else {
		dir = true
	}
	e.sort[field] = dir
	return e
}

//ES Create Index
func (e *ES) CreateIndex(index string) {
	exists, err := e.EsClient.IndexExists(index).Do()
	if err != nil {
		panic(err)
	}
	if !exists {
		createIndex, err := e.EsClient.CreateIndex(index).Do()
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
			panic("create index failed!")
		}
	}
}

//ES Delete Index
func (e *ES) DeleteIndex(index string) {
	exists, err := e.EsClient.IndexExists(index).Do()
	if err != nil {
		panic(err)
	}
	if exists {
		_, err = e.EsClient.DeleteIndex(index).Do()
		if err != nil {
			panic(err)
		}
	} else {
		panic("index: " + index + " not exist.")
	}
}

//ES Get Response
func (e *ES) Search() (ESResponse, error) {
	var resp ESResponse

	// Create service and use query, aggregations, filter
	search := e.EsClient.Search().
		Index(e.esIndex).
		Type(e.esType).
		Query(e.boolQuery).
		Size(e.size).
		From(e.form)

	// Add aggregation
	if len(e.aggregations) > 0 {
		for k, v := range e.aggregations {
			search = search.Aggregation(k, v)
		}
	}

	// Add SortBy
	if len(e.sort) > 0 {
		for k, v := range e.sort {
			search = search.Sort(k, v)
		}
	}

	// Execute query
	sr, err := search.Do()
	if err != nil {
		panic(err)
	}

	// Decode response
	rs, aggs, err := e.decodeResponse(sr)
	if err != nil {
		return resp, err
	}
	resp.Logs = rs
	resp.Total = sr.Hits.TotalHits
	resp.Aggs = aggs
	return resp, nil
}

// DecodeLogs takes a search result and deserialize the response.
func (e *ES) decodeResponse(res *elastic.SearchResult) ([]*LogFormat, elastic.Aggregations, error) {
	if res == nil || res.TotalHits() == 0 {
		return nil, nil, nil
	}

	var rss []*LogFormat
	for _, hit := range res.Hits.Hits {
		rs := new(LogFormat)
		if err := json.Unmarshal(*hit.Source, rs); err != nil {
			return nil, nil, err
		}
		rss = append(rss, rs)
	}
	return rss, res.Aggregations, nil
}
