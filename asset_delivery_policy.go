package ams

import "time"

type AssetDeliveryPolicy struct {
	ID                         string `json:"Id"`
	Name                       string
	AssetDeliveryProtocol      int
	AssetDeliveryPolicyType    int
	AssetDeliveryConfiguration string
	Created                    time.Time
	LastModified               time.Time
}
