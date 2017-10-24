package blob

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
)

type BlockList struct {
	XMLName xml.Name `xml:"BlockList"`
	Blocks  []string `xml:"Latest"`
}

func buildBlockID(id int) string {
	s := fmt.Sprintf("block-id-%v", id)
	return base64.StdEncoding.EncodeToString([]byte(s))
}
