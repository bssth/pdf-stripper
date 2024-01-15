package rules

// Sample rule set

const PdfWidth = 800

var Set1 = []*Rule{
	{
		X1:   0,
		Y1:   0,
		X2:   1000,
		Y2:   150,
		Type: TypeOnFirst,
	},
	{
		X1:   0,
		Y1:   PdfWidth - 150,
		X2:   1000,
		Y2:   PdfWidth,
		Type: TypeOnFirst,
	},
	{
		X1:   0,
		Y1:   PdfWidth - 10,
		X2:   1000,
		Y2:   PdfWidth,
		Type: TypeAllButFirst,
	},
	{
		X1:   0,
		Y1:   0,
		X2:   1000,
		Y2:   50,
		Type: TypeAllButFirst,
	},
	{
		X1:   0,
		Y1:   400,
		X2:   1000,
		Y2:   0,
		Type: TypeLast,
	},
}

var Set2 = []*Rule{
	{
		X1:       0,
		Y1:       -100,
		X2:       1000,
		Y2:       PdfWidth / 4,
		Type:     TypeOnFirst,
		OnlyText: false,
		DelPage:  0,
	},
	{
		X1:       0,
		Y1:       PdfWidth - 150,
		X2:       1000,
		Y2:       PdfWidth,
		Type:     TypeOnFirst,
		OnlyText: false,
		DelPage:  0,
	},
	{
		X1:       0,
		Y1:       PdfWidth - 10,
		X2:       1000,
		Y2:       PdfWidth,
		Type:     TypeAllButFirst,
		OnlyText: false,
		DelPage:  0,
	},
	{
		X1:       0,
		Y1:       0,
		X2:       1000,
		Y2:       50,
		Type:     TypeAllButFirst,
		OnlyText: false,
		DelPage:  0,
	},
}
