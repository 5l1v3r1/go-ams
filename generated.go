package ams

import (
	"context"
	"net/url"
	"os"
	"time"
)

func (c *Client) CreateAccessPolicy(name string, durationInMinutes float64, permissions int) (*AccessPolicy, error) {
	return c.CreateAccessPolicyWithContext(context.Background(), name, durationInMinutes, permissions)
}
func (c *Client) DeleteAccessPolicy(accessPolicy *AccessPolicy) error {
	return c.DeleteAccessPolicyWithContext(context.Background(), accessPolicy)
}
func (c *Client) GetAsset(assetID string) (*Asset, error) {
	return c.GetAssetWithContext(context.Background(), assetID)
}
func (c *Client) GetAssets() ([]Asset, error) {
	return c.GetAssetsWithContext(context.Background())
}
func (c *Client) CreateAsset(name string) (*Asset, error) {
	return c.CreateAssetWithContext(context.Background(), name)
}
func (c *Client) GetAssetFiles(asset *Asset) ([]AssetFile, error) {
	return c.GetAssetFilesWithContext(context.Background(), asset)
}
func (c *Client) CreateAssetFile(assetID, name, mimeType string) (*AssetFile, error) {
	return c.CreateAssetFileWithContext(context.Background(), assetID, name, mimeType)
}
func (c *Client) UpdateAssetFile(assetFile *AssetFile) error {
	return c.UpdateAssetFileWithContext(context.Background(), assetFile)
}
func (c *Client) PutBlob(uploadURL *url.URL, file *os.File) ([]int, error) {
	return c.PutBlobWithContext(context.Background(), uploadURL, file)
}
func (c *Client) PutBlockList(uploadURL *url.URL, blockList []int) error {
	return c.PutBlockListWithContext(context.Background(), uploadURL, blockList)
}
func (c *Client) EncodeAsset(mediaProcessorID, configuration string, asset *Asset) (*Job, error) {
	return c.EncodeAssetWithContext(context.Background(), mediaProcessorID, configuration, asset)
}
func (c *Client) GetOutputMediaAssets(job *Job) ([]Asset, error) {
	return c.GetOutputMediaAssetsWithContext(context.Background(), job)
}
func (c *Client) GetJob(jobID string) (*Job, error) {
	return c.GetJobWithContext(context.Background(), jobID)
}
func (c *Client) WaitJob(job *Job) error {
	return c.WaitJobWithContext(context.Background(), job)
}
func (c *Client) CreateLocator(accessPolicyID, assetID string, startTime time.Time, locatorType int) (*Locator, error) {
	return c.CreateLocatorWithContext(context.Background(), accessPolicyID, assetID, startTime, locatorType)
}
func (c *Client) DeleteLocator(locator *Locator) error {
	return c.DeleteLocatorWithContext(context.Background(), locator)
}
func (c *Client) GetMediaProcessors() ([]MediaProcessor, error) {
	return c.GetMediaProcessorsWithContext(context.Background())
}
