package ams

import (
	"bytes"
	"encoding/xml"
	"testing"
)

func TestEncodeTask(t *testing.T) {
	for _, tc := range []struct {
		data     TaskBody
		expected string
	}{
		{
			data: TaskBody{
				InputAsset:  AssetTag{Asset: "JobInputAsset(0)"},
				OutputAsset: AssetTag{Asset: "JobOutputAsset(0)"},
			},
			expected: `<taskBody><inputAsset>JobInputAsset(0)</inputAsset><outputAsset>JobOutputAsset(0)</outputAsset></taskBody>`,
		},
		{
			data: TaskBody{
				InputAsset: AssetTag{Asset: "JobInputAsset(0)"},
				OutputAsset: AssetTag{
					Asset: "JobOutputAsset(0)",
					Name:  "test",
				},
			},
			expected: `<taskBody><inputAsset>JobInputAsset(0)</inputAsset><outputAsset assetName="test">JobOutputAsset(0)</outputAsset></taskBody>`,
		},
	} {
		var b bytes.Buffer
		if err := xml.NewEncoder(&b).Encode(tc.data); err != nil {
			t.Fatal(err)
		}
		if b.String() != tc.expected {
			t.Errorf("unexpected xml. expected: %v, actual: %v", tc.expected, b.String())
		}
	}

}
