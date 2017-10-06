package ams

import "encoding/xml"

type Task struct {
	Name             string `json:"Name"`
	Configuration    string `json:"Configuration"`
	MediaProcessorID string `json:"MediaProcessorId"`
	TaskBody         string `json:"TaskBody"`
}

type TaskBody struct {
	XMLName     xml.Name `xml:"taskBody"`
	InputAsset  AssetTag `xml:"inputAsset"`
	OutputAsset AssetTag `xml:"outputAsset"`
}

type AssetTag struct {
	Asset string `xml:",chardata"`
	Name  string `xml:"assetName,attr,omitempty"`
}

func newTaskBody() *TaskBody {
	return &TaskBody{
		InputAsset:  AssetTag{Asset: jobInputAsset},
		OutputAsset: AssetTag{Asset: jobOutputAsset},
	}
}
