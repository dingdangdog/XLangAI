package sms

import (
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"xlangai/server/internal/repository"
)

func sendTencent(cfg *repository.SmsServiceConfig, phone, code string) error {
	if cfg.APIKey == "" || cfg.SecretKey == "" {
		return fmt.Errorf("tencent sms: missing api_key (secret_id) or secret_key")
	}
	if cfg.SignName == "" || cfg.TemplateCode == "" {
		return fmt.Errorf("tencent sms: missing sign_name or template_code")
	}
	sdkAppID := configString(cfg.Config, "sdk_app_id", "")
	if sdkAppID == "" {
		sdkAppID = configString(cfg.Config, "sdkAppId", "")
	}
	if sdkAppID == "" {
		return fmt.Errorf("tencent sms: config.sdk_app_id required")
	}
	region := cfg.Region
	if region == "" {
		region = "ap-guangzhou"
	}

	credential := common.NewCredential(cfg.APIKey, cfg.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = configString(cfg.Config, "endpoint", "sms.tencentcloudapi.com")

	client, err := sms.NewClient(credential, region, cpf)
	if err != nil {
		return fmt.Errorf("tencent sms client: %w", err)
	}

	e164 := normalizeChinaPhoneE164(phone)

	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = common.StringPtr(sdkAppID)
	req.SignName = common.StringPtr(cfg.SignName)
	req.TemplateId = common.StringPtr(cfg.TemplateCode)
	req.PhoneNumberSet = common.StringPtrs([]string{e164})
	req.TemplateParamSet = common.StringPtrs([]string{code})

	resp, err := client.SendSms(req)
	if err != nil {
		return fmt.Errorf("tencent sms send: %w", err)
	}
	if resp == nil || resp.Response == nil {
		return fmt.Errorf("tencent sms: empty response")
	}
	if len(resp.Response.SendStatusSet) > 0 {
		st := resp.Response.SendStatusSet[0]
		if st != nil && st.Code != nil && *st.Code != "Ok" {
			msg := ""
			if st.Message != nil {
				msg = *st.Message
			}
			return fmt.Errorf("tencent sms rejected: code=%s message=%s", *st.Code, msg)
		}
	}
	return nil
}

func normalizeChinaPhoneE164(phone string) string {
	p := strings.TrimSpace(phone)
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")
	if strings.HasPrefix(p, "+") {
		return p
	}
	if strings.HasPrefix(p, "00") {
		return "+" + strings.TrimPrefix(p, "00")
	}
	if strings.HasPrefix(p, "86") && len(p) > 11 {
		return "+" + p
	}
	return "+86" + p
}
