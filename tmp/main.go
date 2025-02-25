package main

import (
	"fmt"
	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	// Examine chart.BarChart structure
	fmt.Println("BarChart fields:")
	bc := chart.BarChart{}
	fmt.Printf("- TitleStyle type: %T\n", bc.TitleStyle)
	
	// Examine XAxis structure
	fmt.Println("\nXAxis structure:")
	xa := chart.XAxis{}
	fmt.Printf("- Name: %s\n", xa.Name)
	fmt.Printf("- NameStyle type: %T\n", xa.NameStyle)
	fmt.Printf("- Style type: %T\n", xa.Style)
	
	// Examine Style structure
	fmt.Println("\nStyle structure:")
	s := chart.Style{}
	fmt.Printf("- FontSize field: %d\n", s.FontSize)
	
	// Dump fields for debugging
	fmt.Println("\nDumping Style fields:")
	fmt.Printf("%+v\n", s)
}
