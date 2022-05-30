package internal

import (
	"github.com/kazekiddo/maroto/internal/fpdf"
	"github.com/kazekiddo/maroto/pkg/consts"
	"github.com/kazekiddo/maroto/pkg/props"
)

// Text is the abstraction which deals of how to add text inside PDF.
type Text interface {
	Add(text string, cell Cell, textProp props.Text)
	GetLinesQuantity(text string, fontFamily props.Text, colWidth float64) int
}

type text struct {
	pdf  fpdf.Fpdf
	math Math
	font Font
}

// NewText create a Text.
func NewText(pdf fpdf.Fpdf, math Math, font Font) *text {
	return &text{
		pdf,
		math,
		font,
	}
}

func show_strlen(s string) int {
	sl := 0
	rs := []rune(s)
	for _, r := range rs {
		rint := int(r)
		if rint < 128 {
			sl++
		} else {
			sl += 2
		}
	}
	return sl
}

func show_substr(s string, l int) (string, string) {
	if len(s) <= l {
		return s, ""
	}
	sr, ss, sl, rl, rs := "", "", 0, 0, []rune(s)
	for i, r := range rs {
		rint := int(r)
		if rint < 128 {
			rl = 1
		} else {
			rl = 2
		}

		if sl+rl > l {
			sr = string(rs[i:])
			break
		}
		sl += rl
		ss += string(r)
	}
	if ss == "" {
		return sr, ss
	}
	return ss, sr
}

func (s *text) getLineMaxStr(cellWidth float64, unicodeText string) (string, string) {
	lineText := ""
	for i, v := range unicodeText {
		if s.pdf.GetStringWidth(lineText+string(v)) > cellWidth {
			return lineText, string(unicodeText[i:])
		} else {
			lineText += string(v)
		}
	}
	return lineText, ""
}

// 自适应宽带切割字符串
func (s *text) splitText(cellWidth float64, unicodeText string) []string {
	var sp []string
	// 字体位置不存在时，直接返回
	if s.pdf.Err() {
		return []string{unicodeText}
	}
	for {
		ss, sr := s.getLineMaxStr(cellWidth, unicodeText)
		sp = append(sp, ss)
		if sr == "" {
			break
		}
		unicodeText = sr
	}
	return sp
}

// Add a text inside a cell.
func (s *text) Add(text string, cell Cell, textProp props.Text) {
	s.font.SetFont(textProp.Family, textProp.Style, textProp.Size)

	originalColor := s.font.GetColor()
	s.font.SetColor(textProp.Color)

	// duplicated
	_, _, fontSize := s.font.GetFont()
	fontHeight := fontSize / s.font.GetScaleFactor()

	cell.Y += fontHeight
	//cell.Y += s.pdf.GetY()

	// Apply Unicode before calc spaces
	unicodeText := s.textToUnicode(text, textProp)
	stringWidth := s.pdf.GetStringWidth(unicodeText)
	//words := strings.Split(unicodeText, " ")
	words := s.splitText(cell.Width, unicodeText)
	accumulateOffsetY := 0.0
	// If should add one line
	if stringWidth < cell.Width || textProp.Extrapolate {
		s.addLine(textProp, cell.X, cell.Width, cell.Y, stringWidth, unicodeText)
	} else {
		lines := s.getLines(words, cell.Width)

		for index, line := range lines {
			lineWidth := s.pdf.GetStringWidth(line)
			_, _, fontSize := s.font.GetFont()
			textHeight := fontSize / s.font.GetScaleFactor()

			s.addLine(textProp, cell.X, cell.Width, cell.Y+float64(index)*textHeight+accumulateOffsetY, lineWidth, line)
			accumulateOffsetY += textProp.VerticalPadding
		}
	}

	s.font.SetColor(originalColor)
}

// GetLinesQuantity retrieve the quantity of lines which a text will occupy to avoid that text to extrapolate a cell.
func (s *text) GetLinesQuantity(text string, textProp props.Text, colWidth float64) int {
	//translator := s.pdf.UnicodeTranslatorFromDescriptor("")
	s.font.SetFont(textProp.Family, textProp.Style, textProp.Size)

	// Apply Unicode.
	//textTranslated := translator(text)
	unicodeText := s.textToUnicode(text, textProp)

	stringWidth := s.pdf.GetStringWidth(unicodeText)
	//words := strings.Split(textTranslated, " ")
	words := s.splitText(colWidth, unicodeText)

	// If should add one line.
	if stringWidth < colWidth || textProp.Extrapolate || len(words) == 1 {
		return 1
	}

	lines := s.getLines(words, colWidth)
	return len(lines)
}

func (s *text) getLines(words []string, colWidth float64) []string {
	currentlySize := 0.0
	actualLine := 0

	lines := []string{}
	lines = append(lines, "")

	for _, word := range words {
		if s.pdf.GetStringWidth(word+" ")+currentlySize < colWidth {
			lines[actualLine] = lines[actualLine] + word + " "
			currentlySize += s.pdf.GetStringWidth(word + " ")
		} else {
			lines = append(lines, "")
			actualLine++
			lines[actualLine] = lines[actualLine] + word + " "
			currentlySize = s.pdf.GetStringWidth(word + " ")
		}
	}

	return lines
}

func (s *text) addLine(textProp props.Text, xColOffset, colWidth, yColOffset, textWidth float64, text string) {
	left, top, _, _ := s.pdf.GetMargins()

	if textProp.Align == consts.Left {
		s.pdf.Text(xColOffset+left, yColOffset+top, text)
		return
	}

	var modifier float64 = 2

	if textProp.Align == consts.Right {
		modifier = 1
	}

	dx := (colWidth - textWidth) / modifier

	s.pdf.Text(dx+xColOffset+left, yColOffset+top, text)
}

func (s *text) textToUnicode(txt string, props props.Text) string {
	if props.Family == consts.Arial ||
		props.Family == consts.Helvetica ||
		props.Family == consts.Symbol ||
		props.Family == consts.ZapBats ||
		props.Family == consts.Courier {
		translator := s.pdf.UnicodeTranslatorFromDescriptor("")
		return translator(txt)
	}

	return txt
}
