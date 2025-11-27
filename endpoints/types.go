package endpoints

type Endpoint struct {
	ServerURL                string `json:"serverURL"`
	ProxyURL                 string `json:"proxyURL,omitempty"`
	CertificateAuthorityData string `json:"caData,omitempty"`
	ClientCertificateData    string `json:"-"`
	ClientKeyData            string `json:"-"`
	Token                    string `json:"token,omitempty"`
	Username                 string `json:"username,omitempty"`
	Password                 string `json:"password,omitempty"`
	Debug                    bool   `json:"debug"`
	Insecure                 bool   `json:"insecure"`
	// AWS garbage fields for their proprietary authentication header
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
	AwsRegion    string `json:"awsRegion"`
	AwsService   string `json:"awsService"`
	// AWS time field, testing only
	AwsTime string `json:"awsTime"`
}

// HasCA returns whether the configuration has a certificate authority or not.
func (ep *Endpoint) HasCA() bool {
	return len(ep.CertificateAuthorityData) > 0
}

// HasBasicAuth returns whether the configuration has basic authentication or not.
func (ep *Endpoint) HasBasicAuth() bool {
	return len(ep.Password) != 0
}

// HasTokenAuth returns whether the configuration has token authentication or not.
func (ep *Endpoint) HasTokenAuth() bool {
	return len(ep.Token) != 0
}

// HasCertAuth returns whether the configuration has certificate authentication or not.
func (ep *Endpoint) HasCertAuth() bool {
	return len(ep.ClientCertificateData) != 0 && len(ep.ClientKeyData) != 0
}

// HasAwsAuth returns whether the configuration has AWS authentication or not.
func (ep *Endpoint) HasAwsAuth() bool {
	return len(ep.AwsAccessKey) != 0 &&
		len(ep.AwsSecretKey) != 0 &&
		len(ep.AwsRegion) != 0 &&
		len(ep.AwsSecretKey) != 0
}
