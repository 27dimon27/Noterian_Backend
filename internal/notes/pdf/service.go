package pdf

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	_ "embed"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/jung-kurt/gofpdf/v2"
)

//go:embed fonts/arial.ttf
var arialRegularTTF []byte

//go:embed fonts/arialbd.ttf
var arialBoldTTF []byte

//go:embed fonts/ariali.ttf
var arialItalicTTF []byte

//go:embed fonts/arialbi.ttf
var arialBoldItalicTTF []byte

//go:embed fonts/dejavusans.ttf
var dejavuSansTTF []byte

func registerPDFFonts(pdf *gofpdf.Fpdf) {
	pdf.AddUTF8FontFromBytes("Arial", "", arialRegularTTF)
	pdf.AddUTF8FontFromBytes("Arial", "B", arialBoldTTF)
	pdf.AddUTF8FontFromBytes("Arial", "I", arialItalicTTF)
	pdf.AddUTF8FontFromBytes("Arial", "BI", arialBoldItalicTTF)
	pdf.AddUTF8FontFromBytes("DejaVuSans", "", dejavuSansTTF)
}

type NoteContent struct {
	Note       *models.Note
	Blocks     []models.Block
	Formatting map[string]models.BlockFormatting
	Subnotes   map[string]models.Note
	HeaderURL  string
}

var mimeToExt = map[string]string{
	"image/jpeg": "jpg",
	"image/jpg":  "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

func GeneratePDF(content *NoteContent) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	registerPDFFonts(pdf)
	pdf.AddPage()

	addTitle(pdf, content.Note.Title)

	if content.HeaderURL != "" {
		addHeaderImage(pdf, content.HeaderURL)
	}

	for _, block := range content.Blocks {
		if block.BlockTypeID == 5 {
			block.Content = content.Subnotes[block.ID.String()].Title
		}
		addBlock(pdf, block, content.Formatting[block.ID.String()])
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &buf, nil
}

func addTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 10, title, "", "L", false)
	pdf.Ln(5)
}

func addHeaderImage(pdf *gofpdf.Fpdf, imageURL string) {
	if imageURL == "" {
		return
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body in addHeaderImage: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return
	}

	contentType := resp.Header.Get("Content-Type")
	contentType = strings.Split(contentType, ";")[0]

	extension, ok := mimeToExt[contentType]
	if !ok {
		return
	}

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	opts := gofpdf.ImageOptions{ImageType: extension}
	imgInfo := pdf.RegisterImageOptionsReader(imageURL, opts, bytes.NewReader(imgData))
	if imgInfo != nil {
		placeImage(pdf, imageURL, imgInfo, opts)
		return
	}

	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.MultiCell(0, 5, "Header Image: "+imageURL, "", "L", false)
	pdf.Ln(3)
}

func addBlock(pdf *gofpdf.Fpdf, block models.Block, formatting models.BlockFormatting) {
	// 1=text, 2=image, 3=code, 4=quote, 5=subnote, 6=music, 7=video
	switch block.BlockTypeID {
	case 1:
		addTextBlock(pdf, block.Content, formatting)
	case 2:
		addImageAttachment(pdf, block.Content)
	case 3:
		addCodeBlock(pdf, block.Content)
	case 4:
		addQuoteBlock(pdf, block.Content)
	case 5:
		addSubnoteBlock(pdf, block.Content)
	default:
		addTextBlock(pdf, block.Content, formatting)
	}
}

func addTextBlock(pdf *gofpdf.Fpdf, content string, formatting models.BlockFormatting) {
	if content == "" {
		pdf.Ln(9)
		return
	}

	content = normalizeText(content)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	if len(formatting.Ranges) > 0 {
		addFormattedText(pdf, content, formatting.Ranges)
	} else {
		trimmed := strings.Trim(content, "\n")
		if trimmed != "" {
			pdf.MultiCell(0, 6, trimmed, "", "L", false)
		}
	}

	pdf.Ln(3)
}

func addFormattedText(pdf *gofpdf.Fpdf, content string, ranges []models.FormattingRange) {
	type segment struct {
		text      string
		startPos  int
		endPos    int
		bold      bool
		italic    bool
		underline bool
		textAlign int
	}

	var segments []segment
	lastPos := 0

	for _, r := range ranges {
		if r.StartPos > lastPos {
			segments = append(segments, segment{
				text:      content[lastPos:r.StartPos],
				startPos:  lastPos,
				endPos:    r.StartPos,
				bold:      false,
				italic:    false,
				underline: false,
			})
		}

		text := ""
		if r.EndPos <= len(content) {
			text = content[r.StartPos:r.EndPos]
		} else {
			text = content[r.StartPos:]
		}

		segments = append(segments, segment{
			text:      text,
			startPos:  r.StartPos,
			endPos:    r.EndPos,
			bold:      r.Bold != nil && *r.Bold,
			italic:    r.Italic != nil && *r.Italic,
			underline: r.Underline != nil && *r.Underline,
			textAlign: getTextAlign(r.TextAlign),
		})

		lastPos = r.EndPos
	}

	if lastPos < len(content) {
		segments = append(segments, segment{
			text:      content[lastPos:],
			startPos:  lastPos,
			endPos:    len(content),
			bold:      false,
			italic:    false,
			underline: false,
		})
	}

	for _, seg := range segments {
		txt := strings.Trim(seg.text, "\n")
		if txt == "" {
			continue
		}

		style := ""
		if seg.bold {
			style += "B"
		}
		if seg.italic {
			style += "I"
		}
		if seg.underline {
			style += "U"
		}

		pdf.SetFont("Arial", style, 11)

		pdf.Write(6, txt)
	}
	pdf.Write(6, "\n")
}

func normalizeText(s string) string {
	if s == "" {
		return s
	}

	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "–", "-")
	s = strings.ReplaceAll(s, "\r\n", "\n")

	return s
}

func fitImageSize(info *gofpdf.ImageInfoType, maxWidth, maxHeight float64) (float64, float64) {
	imgWidth := info.Width()
	imgHeight := info.Height()
	if imgWidth <= 0 || imgHeight <= 0 {
		return maxWidth, maxHeight
	}

	width := imgWidth
	height := imgHeight
	if width > maxWidth {
		height = height * (maxWidth / width)
		width = maxWidth
	}
	if height > maxHeight {
		width = width * (maxHeight / height)
		height = maxHeight
	}
	return width, height
}

func placeImage(pdf *gofpdf.Fpdf, imageName string, info *gofpdf.ImageInfoType, opts gofpdf.ImageOptions) {
	pageWidth, pageHeight := pdf.GetPageSize()
	left, top, right, bottom := pdf.GetMargins()
	maxWidth := pageWidth - left - right
	maxHeight := pageHeight - top - bottom

	width, height := fitImageSize(info, maxWidth, maxHeight)
	if width <= 0 || height <= 0 {
		return
	}

	remainingHeight := pageHeight - pdf.GetY() - bottom
	if remainingHeight < height+5 {
		pdf.AddPage()
	}

	pdf.ImageOptions(imageName, left, pdf.GetY(), width, height, false, opts, 0, "")
	pdf.Ln(height + 5)
}

func addImageAttachment(pdf *gofpdf.Fpdf, imageURL string) {
	if imageURL == "" {
		return
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body in addImageAttachment: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return
	}

	contentType := resp.Header.Get("Content-Type")
	contentType = strings.Split(contentType, ";")[0]

	extension, ok := mimeToExt[contentType]
	if !ok {
		return
	}

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	opts := gofpdf.ImageOptions{ImageType: extension}
	imgInfo := pdf.RegisterImageOptionsReader(imageURL, opts, bytes.NewReader(imgData))
	if imgInfo != nil {
		placeImage(pdf, imageURL, imgInfo, opts)
		return
	}

	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(50, 100, 150)
	pdf.MultiCell(0, 8, "Image: "+imageURL, "", "L", false)
	pdf.Ln(5)
}

func addCodeBlock(pdf *gofpdf.Fpdf, content string) {
	if content == "" {
		return
	}

	pdf.SetFont("Courier", "", 10)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 6, content, "", "L", true)
	pdf.Ln(3)
}

func addQuoteBlock(pdf *gofpdf.Fpdf, content string) {
	if content == "" {
		return
	}

	pdf.SetFont("Arial", "I", 11)
	pdf.SetTextColor(80, 80, 80)
	startX := pdf.GetX()
	pdf.SetX(startX + 5)
	pdf.MultiCell(0, 6, content, "", "L", false)
	pdf.SetX(startX)
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(3)
}

func addSubnoteBlock(pdf *gofpdf.Fpdf, subnoteTitle string) {
	label := subnoteTitle
	if label == "" {
		label = "Subnote"
	}

	pdf.SetFont("DejaVuSans", "", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.Write(6, "    ↳ ")
	pdf.SetFont("Arial", "B", 11)
	pdf.Write(6, label)
	pdf.Write(10, "\n")
}

func getTextAlign(alignPtr *int) int {
	if alignPtr == nil {
		return 0
	}
	return *alignPtr
}
