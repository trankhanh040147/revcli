package message

import (
	"path/filepath"
	"strings"
)

type Attachment struct {
	FilePath string
	FileName string
	MimeType string
	Content  []byte
}

func (a Attachment) IsText() bool  { return strings.HasPrefix(a.MimeType, "text/") }
func (a Attachment) IsImage() bool { return strings.HasPrefix(a.MimeType, "image/") }

func NewTextAttachment(filePath, content string) Attachment {
	return Attachment{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
		MimeType: "text/plain",
		Content:  []byte(content),
	}
}
