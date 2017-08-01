package ams

import "fmt"

type Credentials struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	ExpiresOn    string `json:"expires_on"`
	ExtExpiresIn string `json:"ext_expires_in"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}

func (c *Credentials) Token() string {
	return fmt.Sprintf("%s %s", c.TokenType, c.AccessToken)
}
