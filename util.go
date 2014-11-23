package main

import (
	"github.com/creamdog/goamz/logs"
	"strings"
	"fmt"
)

func filter(list []*logs.Event, f func(*logs.Event) bool) []*logs.Event {
	var filteredList = make([]*logs.Event, 0)
	for _, item := range list {
		if f(item) {
			filteredList = append(filteredList, item)
		}
	}
	return filteredList
}

func transform(list []*logs.Event, t func(*logs.Event) *logs.Event) []*logs.Event {
	for _, item := range list {
		item = t(item)
	}
	return list
}


type Table struct {
	Rows [][]string
	ColumnSizes []int
}

func NewTable() *Table {
	t := &Table{make([][]string, 0), make([]int, 0)}
	return t
}	

func (t *Table) Print() {

	separator := func() {
		for i:=0;i<len(t.ColumnSizes);i++ {
			if i == 0 {
				fmt.Printf("+")
			}
			fmt.Printf("-%s-+", strings.Repeat("-", t.ColumnSizes[i]))
		}
		fmt.Printf("\n")		
	}

	separator()

	for _, row := range t.Rows {
		for index, value := range row {
			if index == 0 {
				fmt.Printf("|")
			}
			filler := strings.Repeat(" ", t.ColumnSizes[index] - len(value))
			fmt.Printf(" %s%s |", value, filler)
		}
		fmt.Printf("\n")
		separator()
	}
}

func (t *Table) AddRow(columns ... string) {
	row := make([]string, len(columns))
	copy(row, columns)
	t.Rows = append(t.Rows, row)
	t.recalc()
}

func (t *Table) recalc() {

	numColumns := 0
	for _, row := range t.Rows {
		if numColumns < len(row) {
			numColumns = len(row)
		}	
	}
	t.ColumnSizes = make([]int, numColumns)

	for _, row := range t.Rows {
		for index, value := range row {
			if t.ColumnSizes[index] < len(value) {
				t.ColumnSizes[index] = len(value)
			}
		}
	}

}