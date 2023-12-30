package model

type TestStructure struct {
	Name         string
	Method       string
	URL          string
	Data         interface{}
	ExpectedData interface{}
	PositiveTest bool
	WhatError    error
	Mock         []func([]interface{}, []interface{})
	MockData     [][]interface{}
	QueryParams  map[string]interface{}
	SkipFields   []string
	SkipRoot     string
}
