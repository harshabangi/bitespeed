package util

import "fmt"

type Void struct{}

var VoidValue Void

func KeyExists(k string, m map[string]Void) bool {
	_, exists := m[k]
	return exists
}

type QueryParams struct {
	ColumnNumber int
	Columns      []string
	PlaceHolders []string
	Params       []interface{}
}

func NewQueryParams() QueryParams {
	return QueryParams{
		ColumnNumber: 1,
		Columns:      []string{},
		Params:       []interface{}{},
		PlaceHolders: []string{},
	}
}

func (q *QueryParams) AddParam(columnName string, columnValue interface{}) {
	q.Columns = append(q.Columns, columnName)
	q.Params = append(q.Params, columnValue)
	q.PlaceHolders = append(q.PlaceHolders, fmt.Sprintf("$%d", q.ColumnNumber))
	q.ColumnNumber++
}
