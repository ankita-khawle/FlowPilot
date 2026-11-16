package document

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	pdfreader "github.com/ledongthuc/pdf"
)

const (
	maxDocumentPages       = 2000
	maxDOCXDocumentXMLSize = 20 << 20 // 20 MB uncompressed XML
	maxDOCXEntries         = 1000
)

var (
	ErrUnsupportedFormat = errors.New("unsupported document format")
	ErrInvalidDocument   = errors.New("invalid document")
	ErrEmptyDocument     = errors.New("document contains no extractable text")
	ErrDocumentTooLarge  = errors.New("document exceeds extraction limits")
)

type Parser interface {
	Parse(
		ctx context.Context,
		fileName string,
		content []byte,
	) (ParsedContent, error)
}

type DefaultParser struct{}

func NewParser() *DefaultParser {
	return &DefaultParser{}
}

func (p *DefaultParser) Parse(
	ctx context.Context,
	fileName string,
	content []byte,
) (ParsedContent, error) {

	if len(content) == 0 {
		return ParsedContent{}, ErrEmptyDocument
	}

	if err := ctx.Err(); err != nil {
		return ParsedContent{}, err
	}

	extension := strings.ToLower(
		fileExtension(fileName),
	)

	var result ParsedContent
	var err error

	switch extension {
	case ".txt":
		result, err = parseTXT(content)

	case ".pdf":
		result, err = parsePDF(ctx, content)

	case ".docx":
		result, err = parseDOCX(content)

	default:
		return ParsedContent{}, ErrUnsupportedFormat
	}

	if err != nil {
		return ParsedContent{}, err
	}

	result.Text = strings.TrimSpace(result.Text)

	if result.Text == "" {
		return ParsedContent{}, ErrEmptyDocument
	}

	return result, nil
}

func fileExtension(fileName string) string {
	index := strings.LastIndex(fileName, ".")
	if index == -1 {
		return ""
	}

	return fileName[index:]
}

func parseTXT(content []byte) (ParsedContent, error) {
	if !utf8.Valid(content) || bytes.ContainsRune(content, '\x00') {
		return ParsedContent{}, ErrInvalidDocument
	}

	text := strings.TrimSpace(string(content))

	if text == "" {
		return ParsedContent{}, ErrEmptyDocument
	}

	pageNumber := 1

	return ParsedContent{
		Text: text,
		Pages: []PageContent{
			{
				PageNumber: &pageNumber,
				Text:       text,
			},
		},
	}, nil
}

func parsePDF(
	ctx context.Context,
	content []byte,
) (ParsedContent, error) {

	// PDF headers are normally found near the beginning of the file.
	headerLength := len(content)
	if headerLength > 1024 {
		headerLength = 1024
	}

	if !bytes.Contains(content[:headerLength], []byte("%PDF-")) {
		return ParsedContent{}, ErrInvalidDocument
	}

	reader, err := pdfreader.NewReader(
		bytes.NewReader(content),
		int64(len(content)),
	)
	if err != nil {
		return ParsedContent{}, fmt.Errorf(
			"open PDF: %w",
			err,
		)
	}

	pageCount := reader.NumPage()

	if pageCount > maxDocumentPages {
		return ParsedContent{}, ErrDocumentTooLarge
	}

	pages := make([]PageContent, 0, pageCount)
	var fullText strings.Builder

	for pageNumber := 1; pageNumber <= pageCount; pageNumber++ {

		if err := ctx.Err(); err != nil {
			return ParsedContent{}, err
		}

		page := reader.Page(pageNumber)

		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return ParsedContent{}, fmt.Errorf(
				"extract PDF page %d: %w",
				pageNumber,
				err,
			)
		}

		pageText = strings.TrimSpace(pageText)

		if pageText == "" {
			continue
		}

		currentPage := pageNumber

		pages = append(pages, PageContent{
			PageNumber: &currentPage,
			Text:       pageText,
		})

		if fullText.Len() > 0 {
			fullText.WriteString("\n\n")
		}

		fullText.WriteString(pageText)
	}

	if len(pages) == 0 {
		return ParsedContent{}, ErrEmptyDocument
	}

	return ParsedContent{
		Text:  fullText.String(),
		Pages: pages,
	}, nil
}

func parseDOCX(content []byte) (ParsedContent, error) {

	if len(content) < 4 || !bytes.HasPrefix(content, []byte("PK")) {
		return ParsedContent{}, ErrInvalidDocument
	}

	reader, err := zip.NewReader(
		bytes.NewReader(content),
		int64(len(content)),
	)
	if err != nil {
		return ParsedContent{}, fmt.Errorf(
			"open DOCX archive: %w",
			err,
		)
	}

	if len(reader.File) > maxDOCXEntries {
		return ParsedContent{}, ErrDocumentTooLarge
	}

	var documentFile *zip.File

	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			documentFile = file
			break
		}
	}

	if documentFile == nil {
		return ParsedContent{}, ErrInvalidDocument
	}

	if documentFile.UncompressedSize64 > maxDOCXDocumentXMLSize {
		return ParsedContent{}, ErrDocumentTooLarge
	}

	file, err := documentFile.Open()
	if err != nil {
		return ParsedContent{}, fmt.Errorf(
			"open DOCX document XML: %w",
			err,
		)
	}
	defer file.Close()

	xmlContent, err := io.ReadAll(
		io.LimitReader(file, maxDOCXDocumentXMLSize+1),
	)
	if err != nil {
		return ParsedContent{}, fmt.Errorf(
			"read DOCX document XML: %w",
			err,
		)
	}

	if len(xmlContent) > maxDOCXDocumentXMLSize {
		return ParsedContent{}, ErrDocumentTooLarge
	}

	text, err := extractDOCXParagraphs(xmlContent)
	if err != nil {
		return ParsedContent{}, fmt.Errorf(
			"parse DOCX XML: %w",
			err,
		)
	}

	text = strings.TrimSpace(text)

	if text == "" {
		return ParsedContent{}, ErrEmptyDocument
	}

	return ParsedContent{
		Text: text,
		Pages: []PageContent{
			{
				// DOCX does not provide stable rendered page numbers.
				PageNumber: nil,
				Text:       text,
			},
		},
	}, nil
}

func extractDOCXParagraphs(content []byte) (string, error) {

	decoder := xml.NewDecoder(bytes.NewReader(content))

	var output strings.Builder
	var paragraph strings.Builder

	insideText := false

	for {
		token, err := decoder.Token()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return "", err
		}

		switch element := token.(type) {

		case xml.StartElement:
			if element.Name.Local == "t" {
				insideText = true
			}

		case xml.CharData:
			if insideText {
				paragraph.Write([]byte(element))
			}

		case xml.EndElement:
			switch element.Name.Local {

			case "t":
				insideText = false

			case "p":
				text := strings.TrimSpace(paragraph.String())

				if text != "" {
					if output.Len() > 0 {
						output.WriteString("\n")
					}

					output.WriteString(text)
				}

				paragraph.Reset()
			}
		}
	}

	return output.String(), nil
}
