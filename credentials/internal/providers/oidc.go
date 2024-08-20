package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/credentials-go/credentials/internal/utils"
)

type OIDCCredentialsProvider struct {
	oidcProviderARN     string
	oidcTokenFilePath   string
	roleArn             string
	roleSessionName     string
	durationSeconds     int
	policy              string
	stsRegionId         string
	stsEndpoint         string
	lastUpdateTimestamp int64
	expirationTimestamp int64
	sessionCredentials  *sessionCredentials
	runtime             *utils.Runtime
}

type OIDCCredentialsProviderBuilder struct {
	provider *OIDCCredentialsProvider
}

func NewOIDCCredentialsProviderBuilder() *OIDCCredentialsProviderBuilder {
	return &OIDCCredentialsProviderBuilder{
		provider: &OIDCCredentialsProvider{},
	}
}

func (b *OIDCCredentialsProviderBuilder) WithOIDCProviderARN(oidcProviderArn string) *OIDCCredentialsProviderBuilder {
	b.provider.oidcProviderARN = oidcProviderArn
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithOIDCTokenFilePath(oidcTokenFilePath string) *OIDCCredentialsProviderBuilder {
	b.provider.oidcTokenFilePath = oidcTokenFilePath
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithRoleArn(roleArn string) *OIDCCredentialsProviderBuilder {
	b.provider.roleArn = roleArn
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithRoleSessionName(roleSessionName string) *OIDCCredentialsProviderBuilder {
	b.provider.roleSessionName = roleSessionName
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithDurationSeconds(durationSeconds int) *OIDCCredentialsProviderBuilder {
	b.provider.durationSeconds = durationSeconds
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithStsRegionId(regionId string) *OIDCCredentialsProviderBuilder {
	b.provider.stsRegionId = regionId
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithPolicy(policy string) *OIDCCredentialsProviderBuilder {
	b.provider.policy = policy
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithSTSEndpoint(stsEndpoint string) *OIDCCredentialsProviderBuilder {
	b.provider.stsEndpoint = stsEndpoint
	return b
}

func (b *OIDCCredentialsProviderBuilder) WithRuntime(runtime *utils.Runtime) *OIDCCredentialsProviderBuilder {
	b.provider.runtime = runtime
	return b
}

func (b *OIDCCredentialsProviderBuilder) Build() (provider *OIDCCredentialsProvider, err error) {
	if b.provider.roleSessionName == "" {
		b.provider.roleSessionName = "credentials-go-" + strconv.FormatInt(time.Now().UnixNano()/1000, 10)
	}

	if b.provider.oidcTokenFilePath == "" {
		b.provider.oidcTokenFilePath = os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")
	}

	if b.provider.oidcTokenFilePath == "" {
		err = errors.New("the OIDCTokenFilePath is empty")
		return
	}

	if b.provider.oidcProviderARN == "" {
		b.provider.oidcProviderARN = os.Getenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN")
	}

	if b.provider.oidcProviderARN == "" {
		err = errors.New("the OIDCProviderARN is empty")
		return
	}

	if b.provider.roleArn == "" {
		b.provider.roleArn = os.Getenv("ALIBABA_CLOUD_ROLE_ARN")
	}

	if b.provider.roleArn == "" {
		err = errors.New("the RoleArn is empty")
		return
	}

	if b.provider.durationSeconds == 0 {
		b.provider.durationSeconds = 3600
	}

	if b.provider.durationSeconds < 900 {
		err = errors.New("the Assume Role session duration should be in the range of 15min - max duration seconds")
	}

	if b.provider.stsEndpoint == "" {
		if b.provider.stsRegionId != "" {
			b.provider.stsEndpoint = fmt.Sprintf("sts.%s.aliyuncs.com", b.provider.stsRegionId)
		} else {
			b.provider.stsEndpoint = "sts.aliyuncs.com"
		}
	}

	provider = b.provider
	return
}

func (provider *OIDCCredentialsProvider) getCredentials() (session *sessionCredentials, err error) {
	method := "POST"
	host := provider.stsEndpoint
	queries := make(map[string]string)
	queries["Version"] = "2015-04-01"
	queries["Action"] = "AssumeRoleWithOIDC"
	queries["Format"] = "JSON"
	queries["Timestamp"] = utils.GetTimeInFormatISO8601()

	bodyForm := make(map[string]string)
	bodyForm["RoleArn"] = provider.roleArn
	bodyForm["OIDCProviderArn"] = provider.oidcProviderARN
	token, err := ioutil.ReadFile(provider.oidcTokenFilePath)
	if err != nil {
		return
	}

	bodyForm["OIDCToken"] = string(token)
	if provider.policy != "" {
		bodyForm["Policy"] = provider.policy
	}

	bodyForm["RoleSessionName"] = provider.roleSessionName
	bodyForm["DurationSeconds"] = strconv.Itoa(provider.durationSeconds)

	// caculate signature
	signParams := make(map[string]string)
	for key, value := range queries {
		signParams[key] = value
	}
	for key, value := range bodyForm {
		signParams[key] = value
	}

	querystring := utils.GetURLFormedMap(queries)
	// do request
	httpUrl := fmt.Sprintf("https://%s/?%s", host, querystring)

	body := utils.GetURLFormedMap(bodyForm)

	httpRequest, err := hookNewRequest(http.NewRequest)(method, httpUrl, strings.NewReader(body))
	if err != nil {
		return
	}

	// set headers
	httpRequest.Header["Accept-Encoding"] = []string{"identity"}
	httpRequest.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	httpClient := &http.Client{}

	httpResponse, err := hookDo(httpClient.Do)(httpRequest)
	if err != nil {
		return
	}

	defer httpResponse.Body.Close()

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return
	}

	if httpResponse.StatusCode != http.StatusOK {
		message := "get session token failed: "
		err = errors.New(message + string(responseBody))
		return
	}
	var data assumeRoleResponse
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		err = fmt.Errorf("get oidc sts token err, json.Unmarshal fail: %s", err.Error())
		return
	}
	if data.Credentials == nil {
		err = fmt.Errorf("get oidc sts token err, fail to get credentials")
		return
	}

	if data.Credentials.AccessKeyId == nil || data.Credentials.AccessKeySecret == nil || data.Credentials.SecurityToken == nil {
		err = fmt.Errorf("refresh RoleArn sts token err, fail to get credentials")
		return
	}

	session = &sessionCredentials{
		AccessKeyId:     *data.Credentials.AccessKeyId,
		AccessKeySecret: *data.Credentials.AccessKeySecret,
		SecurityToken:   *data.Credentials.SecurityToken,
		Expiration:      *data.Credentials.Expiration,
	}
	return
}

func (provider *OIDCCredentialsProvider) needUpdateCredential() (result bool) {
	if provider.expirationTimestamp == 0 {
		return true
	}

	return provider.expirationTimestamp-time.Now().Unix() <= 180
}

func (provider *OIDCCredentialsProvider) GetCredentials() (cc *Credentials, err error) {
	if provider.sessionCredentials == nil || provider.needUpdateCredential() {
		sessionCredentials, err1 := provider.getCredentials()
		if err1 != nil {
			return nil, err1
		}

		provider.sessionCredentials = sessionCredentials
		expirationTime, err2 := time.Parse("2006-01-02T15:04:05Z", sessionCredentials.Expiration)
		if err2 != nil {
			return nil, err2
		}

		provider.lastUpdateTimestamp = time.Now().Unix()
		provider.expirationTimestamp = expirationTime.Unix()
	}

	cc = &Credentials{
		AccessKeyId:     provider.sessionCredentials.AccessKeyId,
		AccessKeySecret: provider.sessionCredentials.AccessKeySecret,
		SecurityToken:   provider.sessionCredentials.SecurityToken,
		ProviderName:    provider.GetProviderName(),
	}
	return
}

func (provider *OIDCCredentialsProvider) GetProviderName() string {
	return "oidc"
}