package blob

import (
	"encoding/xml"
	"testing"
)

func TestBlockList_MarshalXML(t *testing.T) {
	blockList := new(BlockList)
	blockList.Blocks = []string{"1", "2", "3", "4"}
	expected := `<BlockList><Latest>1</Latest><Latest>2</Latest><Latest>3</Latest><Latest>4</Latest></BlockList>`
	b, err := xml.Marshal(blockList)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if got != expected {
		t.Errorf("unexpected xml. expected: %v, got: %v", expected, got)
	}
}
