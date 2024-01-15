package rules

import (
	"github.com/unidoc/unipdf/v3/core"
	"log"
)

// NeedRemoveAt checks if object at X and Y on page P is going to be deleted because of deletion rules
func NeedRemoveAt(rules []*Rule, x float64, y float64, page int, pagesCnt int) bool {
	for _, rule := range rules {
		// Check if position is in square
		if x < rule.X1 || x > rule.X2 {
			continue
		}
		if y < rule.Y1 || y > rule.Y2 {
			continue
		}

		// If rule is only for first page
		if rule.Type == TypeOnFirst && page != 1 {
			continue
		}
		// ...or every page but first
		if rule.Type == TypeAllButFirst && page == 1 {
			continue
		}
		// ...or only last one
		if rule.Type == TypeLast && page == pagesCnt {
			log.Println(x, y, rule.X1, rule.X2)
		}
		if rule.Type == TypeLast && page != pagesCnt {
			continue
		}
		return true
	}

	return false
}

// ToFloat safely converts PDF value to float
func ToFloat(obj core.PdfObject) float64 {
	if f, ok := obj.(*core.PdfObjectFloat); ok {
		return float64(*f)
	}
	if f, ok := obj.(*core.PdfObjectInteger); ok {
		return float64(*f)
	}

	return 0
}
