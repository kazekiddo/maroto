package internal

import (
	"github.com/kazekiddo/maroto/pkg/color"
	"github.com/kazekiddo/maroto/pkg/consts"
	"github.com/kazekiddo/maroto/pkg/props"
)

const (
	lineHeight = 1.0
)

// MarotoGridPart is the abstraction to deal with the gris system inside the table list.
type MarotoGridPart interface {
	// Grid System.
	Row(height float64, closure func())
	Col(width uint, closure func())
	ColWithMaxGridSum(width uint, closure func(), maxGridSum float64)
	ColSpace(width uint)

	// Helpers.
	SetBackgroundColor(color color.Color)
	GetCurrentOffset() float64
	GetPageSize() (width float64, height float64)
	GetPageMargins() (left float64, top float64, right float64, bottom float64)

	// Outside Col/Row Components.
	Line(spaceHeight float64, line ...props.Line)

	// Inside Col/Row Components.
	Text(text string, prop ...props.Text)

	//
	SetBorder(on bool)
}

// TableList is the abstraction to create a table with header and contents.
type TableList interface {
	Create(header []string, contents [][]string, defaultFontFamily string, prop ...props.TableList) (line int)
	BindGrid(part MarotoGridPart)
}

type tableList struct {
	pdf  MarotoGridPart
	text Text
	font Font
}

// NewTableList create a TableList.
func NewTableList(text Text, font Font) *tableList {
	return &tableList{
		text: text,
		font: font,
	}
}

// BindGrid bind the grid system to TableList.
func (s *tableList) BindGrid(pdf MarotoGridPart) {
	s.pdf = pdf
}

// Create method creates a header section with a list of strings and
// create many rows with contents.
func (s *tableList) Create(header []string, contents [][]string, defaultFontFamily string, prop ...props.TableList) (line int) {
	if len(header) == 0 {
		return
	}

	if len(contents) == 0 {
		return
	}

	tableProp := props.TableList{}

	if len(prop) > 0 {
		tableProp = prop[0]
	}
	maxGridSum := s.customizeMaxGridSum(tableProp.ContentProp)
	if maxGridSum < float64(len(header)) {
		maxGridSum = float64(len(header))
	}

	tableProp.MakeValid(header, defaultFontFamily)
	headerHeight, linesHeight := s.calcLinesHeight(header, tableProp.HeaderProp, tableProp.Align)

	// Draw header.
	s.pdf.Row(headerHeight+1, func() {
		for i, h := range header {
			hs := h
			top := (headerHeight - linesHeight[i]) / 2
			s.pdf.ColWithMaxGridSum(tableProp.HeaderProp.GridSizes[i], func() {
				reason := hs
				s.pdf.Text(reason, tableProp.HeaderProp.ToTextProp(tableProp.Aligns[i], top, false, 0.0))
			}, maxGridSum)
		}
	})

	// Define space between header and contents.
	s.pdf.Row(tableProp.HeaderContentSpace, func() {
		s.pdf.ColSpace(0)
	})

	line = len(contents)
	// Draw contents.
	for index, content := range contents {
		contentHeight, linesHeight := s.calcLinesHeight(content, tableProp.ContentProp, tableProp.Align)
		contentHeightPadded := contentHeight + tableProp.VerticalContentPadding

		if tableProp.AlternatedBackground != nil && index%2 == 0 {
			s.pdf.SetBackgroundColor(*tableProp.AlternatedBackground)
		}
		_, pageHeight := s.pdf.GetPageSize()
		_, top, _, bottom := s.pdf.GetPageMargins()
		offsetY := s.pdf.GetCurrentOffset()
		maxOffsetPage := pageHeight - bottom - top
		if offsetY > (maxOffsetPage - contentHeight - 1 - 7) {
			s.pdf.SetBorder(false)
			//填充本页剩余空间
			//s.pdf.Row(maxOffsetPage-offsetY, func() {
			//	s.pdf.ColSpace(0)
			//})
			//
			//s.pdf.Row(float64(headerHeight+1) , func() {
			//	s.pdf.ColSpace(0)
			//})
			//s.pdf.SetBorder(true)
			//s.pdf.Row(headerHeight+1, func() {
			//	for i, h := range header {
			//		hs := h
			//		top := (headerHeight - linesHeight[i]) / 2
			//		s.pdf.ColWithMaxGridSum(tableProp.HeaderProp.GridSizes[i], func() {
			//			reason := hs
			//			s.pdf.Text(reason, tableProp.HeaderProp.ToTextProp(tableProp.Aligns[i], top, false, 0.0))
			//		}, maxGridSum)
			//	}
			//})
			line = index
			break
		}
		s.pdf.Row(contentHeightPadded+1, func() {
			for i, c := range content {
				cs := c
				top := (contentHeight - linesHeight[i]) / 2
				s.pdf.ColWithMaxGridSum(tableProp.ContentProp.GridSizes[i], func() {
					s.pdf.Text(cs, tableProp.ContentProp.ToTextProp(tableProp.Aligns[i], top+tableProp.VerticalContentPadding/2.0, false, 0.0))
				}, maxGridSum)
			}
		})

		if tableProp.AlternatedBackground != nil && index%2 == 0 {
			s.pdf.SetBackgroundColor(color.NewWhite())
		}

		if tableProp.Line {
			s.pdf.Line(lineHeight)
		}
	}
	s.pdf.SetBorder(false)
	return
}

func (s *tableList) customizeMaxGridSum(contentProp props.TableListContent) float64 {
	if len(contentProp.GridSizes) == 0 {
		return consts.MaxGridSum
	}
	sizeSum := 0.0
	for _, size := range contentProp.GridSizes {
		sizeSum += float64(size)
	}
	if sizeSum == consts.MaxGridSum {
		sizeSum = consts.MaxGridSum
	}
	return sizeSum
}

func (s *tableList) calcLinesHeight(textList []string, contentProp props.TableListContent, align consts.Align) (float64, []float64) {
	maxLines := 1.0

	left, _, right, _ := s.pdf.GetPageMargins()
	width, _ := s.pdf.GetPageSize()
	usefulWidth := width - left - right

	textProp := contentProp.ToTextProp(align, 0, false, 0.0)
	maxGridSum := s.customizeMaxGridSum(contentProp)

	linesHeight := make([]float64, 0, len(textList))

	for i, text := range textList {
		gridSize := float64(contentProp.GridSizes[i])
		percentSize := gridSize / maxGridSum
		colWidth := usefulWidth * percentSize
		qtdLines := float64(s.text.GetLinesQuantity(text, textProp, colWidth))
		linesHeight = append(linesHeight, qtdLines)
		if qtdLines > maxLines {
			maxLines = qtdLines
		}
	}

	_, _, fontSize := s.font.GetFont()

	// Font size corrected by the scale factor from "mm" inside gofpdf f.k.
	fontHeight := fontSize / s.font.GetScaleFactor()
	for i, f := range linesHeight {
		linesHeight[i] = fontHeight * f
	}

	return fontHeight * maxLines, linesHeight
}
