package rules

const (
	// TypeOnAll remove objects on all pages
	TypeOnAll = 1
	// TypeOnFirst remove objects only on first page
	TypeOnFirst = 2
	// TypeAllButFirst remove objects on all pages in exception to first one
	TypeAllButFirst = 3
	// TypeLast remove objects on very last page
	TypeLast = 4
)

type Rule struct {
	// Position of rectangle vertices. All in this rectangle will be removed by rule
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
	// Type changes rule behaviour (usually set pages to clean)
	Type int
	// Don't remove everything but text strings
	OnlyText bool

	// Delete page instead of objects in rectangle
	DelPage int
}

// GetRuleSet returns some rules set by it`s hardcoded ID
func GetRuleSet(id int) []*Rule {
	switch id {
	case 1:
		return Set1
	case 2:
		return Set2
	}

	return nil
}
