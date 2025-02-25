package main

import (
	"fmt"
	"reflect"
	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	// Create a BarChart instance
	bc := chart.BarChart{}
	
	// Get Type information
	t := reflect.TypeOf(bc)
	
	// List fields and their types
	fmt.Println("BarChart struct fields:")
	fmt.Println("===================================")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("%s: %s\n", field.Name, field.Type)
	}
}
