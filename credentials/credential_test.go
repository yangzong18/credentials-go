package credentials

import (
	"os"
	"testing"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials/internal/utils"
	"github.com/aliyun/credentials-go/credentials/request"
	"github.com/stretchr/testify/assert"
)

var privatekey = `----
this is privatekey`

func TestConfig(t *testing.T) {
	config := new(Config)
	assert.Equal(t, "{\n   \"type\": null,\n   \"access_key_id\": null,\n   \"access_key_secret\": null,\n   \"security_token\": null,\n   \"bearer_token\": null,\n   \"oidc_provider_arn\": null,\n   \"oidc_token\": null,\n   \"role_arn\": null,\n   \"role_session_name\": null,\n   \"role_session_expiration\": null,\n   \"policy\": null,\n   \"external_id\": null,\n   \"sts_endpoint\": null,\n   \"role_name\": null,\n   \"enable_imds_v2\": null,\n   \"disable_imds_v1\": null,\n   \"metadata_token_duration\": null,\n   \"url\": null,\n   \"session_expiration\": null,\n   \"public_key_id\": null,\n   \"private_key_file\": null,\n   \"host\": null,\n   \"timeout\": null,\n   \"connect_timeout\": null,\n   \"proxy\": null,\n   \"inAdvanceScale\": null\n}", config.String())
	assert.Equal(t, "{\n   \"type\": null,\n   \"access_key_id\": null,\n   \"access_key_secret\": null,\n   \"security_token\": null,\n   \"bearer_token\": null,\n   \"oidc_provider_arn\": null,\n   \"oidc_token\": null,\n   \"role_arn\": null,\n   \"role_session_name\": null,\n   \"role_session_expiration\": null,\n   \"policy\": null,\n   \"external_id\": null,\n   \"sts_endpoint\": null,\n   \"role_name\": null,\n   \"enable_imds_v2\": null,\n   \"disable_imds_v1\": null,\n   \"metadata_token_duration\": null,\n   \"url\": null,\n   \"session_expiration\": null,\n   \"public_key_id\": null,\n   \"private_key_file\": null,\n   \"host\": null,\n   \"timeout\": null,\n   \"connect_timeout\": null,\n   \"proxy\": null,\n   \"inAdvanceScale\": null\n}", config.GoString())

	config.SetSTSEndpoint("sts.cn-hangzhou.aliyuncs.com")
	assert.Equal(t, "sts.cn-hangzhou.aliyuncs.com", *config.STSEndpoint)
}

func TestNewCredentialWithNil(t *testing.T) {
	rollback := utils.Memory(EnvVarAccessKeyId, EnvVarAccessKeySecret, "ALIBABA_CLOUD_CLI_PROFILE_DISABLED")
	defer func() {
		rollback()
	}()

	os.Setenv(EnvVarAccessKeyId, "accesskey")
	os.Setenv(EnvVarAccessKeySecret, "accesssecret")

	cred, err := NewCredential(nil)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	os.Unsetenv(EnvVarAccessKeyId)
	os.Unsetenv(EnvVarAccessKeySecret)
	os.Setenv("ALIBABA_CLOUD_CLI_PROFILE_DISABLED", "true")

	cred, err = NewCredential(nil)
	assert.Nil(t, err)
	_, err = cred.GetCredential()
	assert.Contains(t, err.Error(), "unable to get credentials from any of the providers in the chain:")
}

func TestNewCredentialWithAK(t *testing.T) {
	config := new(Config)
	config.SetType("access_key")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the access key id is empty", err.Error())
	assert.Nil(t, cred)

	config.SetAccessKeyId("AccessKeyId")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the access key secret is empty", err.Error())
	assert.Nil(t, cred)

	config.SetAccessKeySecret("AccessKeySecret")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	cm, err := cred.GetCredential()
	assert.Nil(t, err)
	assert.Equal(t, "AccessKeyId", *cm.AccessKeyId)
	assert.Equal(t, "AccessKeySecret", *cm.AccessKeySecret)
	assert.Equal(t, "", *cm.SecurityToken)

	// test deprecated methods
	accessKeyId, err := cred.GetAccessKeyId()
	assert.Nil(t, err)
	assert.Equal(t, "AccessKeyId", *accessKeyId)
	accessKeySecret, err := cred.GetAccessKeySecret()
	assert.Nil(t, err)
	assert.Equal(t, "AccessKeySecret", *accessKeySecret)
	securityToken, err := cred.GetSecurityToken()
	assert.Nil(t, err)
	assert.Equal(t, "", *securityToken)
}

func TestNewCredentialWithSts(t *testing.T) {
	config := new(Config)
	config.SetType("sts")

	config.SetAccessKeyId("")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the access key id is empty", err.Error())
	assert.Nil(t, cred)

	config.SetAccessKeyId("akid")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the access key secret is empty", err.Error())
	assert.Nil(t, cred)

	config.SetAccessKeySecret("aksecret")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the security token is empty", err.Error())
	assert.Nil(t, cred)

	config.SetSecurityToken("SecurityToken")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)
}

func TestNewCredentialWithECSRAMRole(t *testing.T) {
	config := new(Config)
	config.SetType("ecs_ram_role")
	cred, err := NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetRoleName("AccessKeyId")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetEnableIMDSv2(false)
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetDisableIMDSv1(false)
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetEnableIMDSv2(true)
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetDisableIMDSv1(true)
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetEnableIMDSv2(true)
	config.SetMetadataTokenDuration(180)
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)
}

func TestNewCredentialWithRSAKeyPair(t *testing.T) {
	config := new(Config)
	config.SetType("rsa_key_pair")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "PrivateKeyFile cannot be empty", err.Error())
	assert.Nil(t, cred)

	config.SetPrivateKeyFile("test")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "PublicKeyId cannot be empty", err.Error())
	assert.Nil(t, cred)

	config.
		SetPublicKeyId("resource").
		SetPrivateKeyFile("nofile").
		SetSessionExpiration(10).
		SetRoleSessionExpiration(10).
		SetPolicy("").
		SetHost("").
		SetTimeout(10).
		SetConnectTimeout(10).
		SetProxy("")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "InvalidPath: Can not open PrivateKeyFile, err is open nofile:")
	assert.Nil(t, cred)

	file, err := os.Create("./pk.pem")
	assert.Nil(t, err)
	file.WriteString(privatekey)
	file.Close()

	config.SetType("rsa_key_pair").
		SetPublicKeyId("resource").
		SetPrivateKeyFile("./pk.pem")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)
}

func TestNewCredentialWithRAMRoleARN(t *testing.T) {
	config := new(Config)
	config.SetType("ram_role_arn")
	config.SetAccessKeyId("")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the access key id is empty", err.Error())
	assert.Nil(t, cred)

	config.SetAccessKeyId("akid")
	config.SetAccessKeySecret("")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the access key secret is empty", err.Error())
	assert.Nil(t, cred)

	config.SetAccessKeySecret("AccessKeySecret")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the RoleArn is empty", err.Error())
	assert.Nil(t, cred)

	config.SetRoleArn("roleArn")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetRoleSessionName("role_session_name")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	// empty security token should ok
	config.SetSecurityToken("")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	// with sts should ok
	config.SetSecurityToken("securitytoken")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

	config.SetExternalId("externalId")
	config.SetPolicy("policy")
	config.SetRoleSessionExpiration(3600)
	config.SetRoleSessionName("roleSessionName")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)

}

func TestNewCredentialWithBearerToken(t *testing.T) {
	config := new(Config)
	config.SetType("bearer")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "BearerToken cannot be empty", err.Error())
	assert.Nil(t, cred)

	config.SetBearerToken("BearerToken")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)
}

func TestNewCredentialWithOIDC(t *testing.T) {
	config := new(Config)
	// oidc role arn
	config.SetType("oidc_role_arn")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the OIDCTokenFilePath is empty", err.Error())
	assert.Nil(t, cred)

	config.SetOIDCTokenFilePath("oidc_token_file_path_test")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the OIDCProviderARN is empty", err.Error())
	assert.Nil(t, cred)

	config.SetOIDCProviderArn("oidc_provider_arn_test")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "the RoleArn is empty", err.Error())
	assert.Nil(t, cred)

	config.SetRoleArn("role_arn_test")
	cred, err = NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)
	assert.Equal(t, "oidc_provider_arn_test", tea.StringValue(config.OIDCProviderArn))
	assert.Equal(t, "oidc_token_file_path_test", tea.StringValue(config.OIDCTokenFilePath))
	assert.Equal(t, "role_arn_test", tea.StringValue(config.RoleArn))
}

func TestNewCredentialWithCredentialsURI(t *testing.T) {
	config := new(Config)

	config.SetType("credentials_uri").
		SetURLCredential("http://test/")
	cred, err := NewCredential(config)
	assert.Nil(t, err)
	assert.NotNil(t, cred)
	assert.Equal(t, "http://test/", tea.StringValue(config.Url))

	config.SetURLCredential("")
	cred, err = NewCredential(config)
	assert.NotNil(t, err)
	assert.Nil(t, cred)
	assert.Equal(t, "", tea.StringValue(config.Url))
}

func TestNewCredentialWithInvalidType(t *testing.T) {
	config := new(Config)
	config.SetType("sdk")
	cred, err := NewCredential(config)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid type option, support: access_key, sts, bearer, ecs_ram_role, ram_role_arn, rsa_key_pair, oidc_role_arn, credentials_uri", err.Error())
	assert.Nil(t, cred)
}

func Test_doaction(t *testing.T) {
	request := request.NewCommonRequest()
	request.Method = "credential test"
	content, err := doAction(request, nil)
	assert.NotNil(t, err)
	assert.Equal(t, `net/http: invalid method "credential test"`, err.Error())
	assert.Nil(t, content)
	request.Method = "GET"
	request.URL = "http://www.aliyun.com"
	runtime := &utils.Runtime{
		Proxy: "# #%gfdf",
	}
	content, err = doAction(request, runtime)
	assert.Contains(t, err.Error(), `invalid URL escape`)
	assert.NotNil(t, err)
	assert.Nil(t, content)
}
