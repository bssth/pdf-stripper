package main

import (
	"flag"
	rule "github.com/bssth/pdf-stripper/rules"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/model/optimize"
	"log"
	"os"
	"regexp"
	"strings"
)

// textChunk represents some text chunked by parts in PDF file
type textChunk struct {
	font   *model.PdfFont
	strObj *core.PdfObjectString
	val    string
	idx    int
}

func (tc *textChunk) encode() {
	var encoded string
	if font := tc.font; font != nil {
		encodedBytes, numMisses := font.StringToCharcodeBytes(tc.val)
		if numMisses != 0 {
			common.Log.Debug("WARN: some runes could not be encoded.\n\t%s -> %v")
		}
		encoded = string(encodedBytes)
	}

	*tc.strObj = *core.MakeString(encoded)
}

type textChunks struct {
	text   string
	chunks []*textChunk
}

var urlRegex = regexp.MustCompile("(https://|www\\.)[a-z|\\.]+(\\/[^www\\.]+|)")

// Strips urls by pattern in some text chunk
func (tc *textChunks) strip() {
	const replacement = ""
	text := tc.text
	chunks := tc.chunks

	matches := urlRegex.FindAllString(tc.text, -1)
	if matches == nil {
		return
	}

	var chunkOffset int

	for _, search := range matches {
		common.Log.Debug("Match: ", search)
		matchIdx := strings.Index(text, search)
		for currMatchIdx := matchIdx; matchIdx != -1; {
			for i, chunk := range chunks[chunkOffset:] {
				idx, lenChunk := chunk.idx, len(chunk.val)
				if currMatchIdx < idx || currMatchIdx > idx+lenChunk-1 {
					continue
				}
				chunkOffset += i + 1

				start := currMatchIdx - idx
				remaining := len(search) - (lenChunk - start)

				replaceVal := chunk.val[:start] + replacement
				if remaining < 0 {
					replaceVal += chunk.val[lenChunk+remaining:]
					chunkOffset--
				}

				chunk.val = replaceVal
				chunk.encode()

				for j := chunkOffset; remaining > 0; j++ {
					c := chunks[j]
					l := len(c.val)

					if l > remaining {
						c.val = c.val[remaining:]
					} else {
						c.val = ""
						chunkOffset++
					}

					c.encode()
					remaining -= l
				}

				break
			}

			text = text[matchIdx+1:]
			matchIdx = strings.Index(text, search)
			currMatchIdx += matchIdx + 1
		}

		tc.text = strings.Replace(tc.text, search, replacement, -1)
	}
}

func init() {
	err := license.SetMeteredKey(
		os.Getenv("API_KEY"),
	)
	if err != nil {
		panic(err)
	}
}

// Current global rule set to use
var rules []*rule.Rule

func main() {
	input := flag.String("input", "input.pdf", "Input PDF file")
	output := flag.String("output", "output.pdf", "Output file name")
	rulesId := flag.Int("rules", 1, "Rule set to use")
	flag.Parse()

	rules = rule.GetRuleSet(*rulesId)
	err := handleFile(*input, *output)
	if err != nil {
		panic(err)
	}
	log.Println("Successfully created", *output)
}

// Entrypoint for every file to handle
func handleFile(inputPath, outputPath string) error {
	pdfWriter := model.NewPdfWriter()
	pdfReader, f, err := model.NewPdfReaderFromFile(inputPath, nil)
	if err != nil {
		return err
	}
	defer f.Close()

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	for n := 1; n <= numPages; n++ {
		page, err := pdfReader.GetPage(n)
		if err != nil {
			return err
		}

		if box, err := page.GetMediaBox(); err == nil {
			page.CropBox = box
		}

		// @todo maybe use in future
		/*err = handlePageTexts(page)
		if err != nil {
			return err
		}*/

		err = handlePageResources(page, n, numPages)
		if err != nil {
			return err
		}

		err = pdfWriter.AddPage(page)
		if err != nil {
			return err
		}
	}

	fw, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer fw.Close()

	opt := optimize.Options{
		CombineDuplicateStreams:         true,
		CombineIdenticalIndirectObjects: true,
		UseObjectStreams:                true,
		CompressStreams:                 true,
	}
	pdfWriter.SetOptimizer(optimize.New(opt))

	return pdfWriter.Write(fw)
}

// Content-removing handler for every page in document
func handlePageResources(page *model.PdfPage, pageNum int, pagesCnt int) error {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		common.Log.Debug("failed to get page content stream")
	}

	parser := contentstream.NewContentStreamParser(contents)
	ops, err := parser.Parse()
	if err != nil {
		common.Log.Debug("failed to parse content stream")
	}

	var lastX, lastY float64 = 0, 0
	var usedObjectsNames []string

	/**
	PDF format consists of many commands (operands) just like any another imperative script.
	That`s why we need just to iterate over them. Changing position and insert content are
	separate operations, so we use lastX and lastY to remember them.
	*/
	for idx, op := range *ops {
		operand := op.Operand
		// fmt.Printf("%s %v\n", op.Operand, op.Params)

		switch {
		case (operand == "Td" || operand == "m") && len(op.Params) == 2:
			x := rule.ToFloat(op.Params[0])
			if x != 0 {
				lastX = x
			}

			y := rule.ToFloat(op.Params[1])
			if y != 0 {
				lastY = y
			}
		case (operand == "cm" || operand == "Tm") && len(op.Params) >= 6:
			lastX = rule.ToFloat(op.Params[4])
			lastY = rule.ToFloat(op.Params[5])
		case operand == "re" && len(op.Params) >= 2:
			lastX = rule.ToFloat(op.Params[0])
			lastY = rule.ToFloat(op.Params[1])
			// 're' includes coordinates, so we can remove it on place
			if rule.NeedRemoveAt(rules, lastX, lastY, pageNum, pagesCnt) {
				(*ops)[idx] = nil
			}
		case operand == "TJ" || operand == "Tj" || operand == "l" || operand == "c" || operand == "S" || operand == "s" ||
			operand == "f" || operand == "b" || operand == "B" || operand == "*":
			if rule.NeedRemoveAt(rules, lastX, lastY, pageNum, pagesCnt) {
				(*ops)[idx] = nil
			}

		case operand == "Do":
			log.Println(lastX, lastY, pageNum, pagesCnt)
			if rule.NeedRemoveAt(rules, lastX, lastY, pageNum, pagesCnt) {
				(*ops)[idx] = nil
			} else {
				params := op.Params
				imageName := params[0].String()
				usedObjectsNames = append(usedObjectsNames, imageName)
			}
		}
	}

	xObject := page.Resources.XObject
	dict, ok := xObject.(*core.PdfObjectDictionary)
	if ok {
		keys := getKeys(dict)
		for _, k := range keys {
			if exists(k, usedObjectsNames) {
				continue
			}
			name := *core.MakeName(k)
			dict.Remove(name)
			common.Log.Debug("Removing XObject", name)
		}
	}

	return page.SetContentStreams([]string{ops.String()}, core.NewFlateEncoder())
}

// Handler for every page to strip text by patterns
func handlePageTexts(page *model.PdfPage) error {
	contents, err := page.GetAllContentStreams()
	if err != nil {
		return err
	}

	ops, err := contentstream.NewContentStreamParser(contents).Parse()
	if err != nil {
		return err
	}

	var currFont *model.PdfFont
	tc := textChunks{}

	textProcFunc := func(objPtr *core.PdfObject) {
		strObj, ok := core.GetString(*objPtr)
		if !ok {
			common.Log.Debug("Invalid parameter, skipping")
			return
		}

		str := strObj.String()
		if currFont != nil {
			decoded, _, numMisses := currFont.CharcodeBytesToUnicode(strObj.Bytes())
			if numMisses != 0 {
				common.Log.Debug("WARN: some chars could not be decoded.\n\t%v -> %s", strObj.Bytes(), decoded)
			}
			str = decoded
		}

		tc.chunks = append(tc.chunks, &textChunk{
			font:   currFont,
			strObj: strObj,
			val:    str,
			idx:    len(tc.text),
		})
		tc.text += str
	}

	processor := contentstream.NewContentStreamProcessor(*ops)
	processor.AddHandler(contentstream.HandlerConditionEnumAllOperands, "",
		func(op *contentstream.ContentStreamOperation, gs contentstream.GraphicsState, resources *model.PdfPageResources) error {
			switch op.Operand {
			case `Tj`, `'`:
				if len(op.Params) != 1 {
					common.Log.Debug("Invalid: Tj/' with invalid set of parameters - skip")
					return nil
				}
				textProcFunc(&op.Params[0])
			case `''`:
				if len(op.Params) != 3 {
					common.Log.Debug("Invalid: '' with invalid set of parameters - skip")
					return nil
				}
				textProcFunc(&op.Params[3])
			case `TJ`:
				if len(op.Params) != 1 {
					common.Log.Debug("Invalid: TJ with invalid set of parameters - skip")
					return nil
				}
				arr, _ := core.GetArray(op.Params[0])
				for i := range arr.Elements() {
					obj := arr.Get(i)
					textProcFunc(&obj)
					arr.Set(i, obj)
				}
			case "Tf":
				if len(op.Params) != 2 {
					common.Log.Debug("Invalid: Tf with invalid set of parameters - skip")
					return nil
				}

				fname, ok := core.GetName(op.Params[0])
				if !ok || fname == nil {
					common.Log.Debug("ERROR: could not get font name")
					return nil
				}

				fObj, has := resources.GetFontByName(*fname)
				if !has {
					common.Log.Debug("ERROR: font %s not found", fname.String())
					return nil
				}

				pdfFont, err := model.NewPdfFontFromPdfObject(fObj)
				if err != nil {
					common.Log.Debug("ERROR: loading font")
					return nil
				}
				currFont = pdfFont
			}

			return nil
		})

	if err = processor.Process(page.Resources); err != nil {
		return err
	}

	tc.strip()
	return page.SetContentStreams([]string{ops.String()}, core.NewFlateEncoder())
}

func getKeys(dict *core.PdfObjectDictionary) []string {
	var keys []string
	for _, k := range dict.Keys() {
		keys = append(keys, k.String())
	}
	return keys
}

func exists(element string, elements []string) bool {
	for _, el := range elements {
		if element == el {
			return true
		}
	}
	return false
}
