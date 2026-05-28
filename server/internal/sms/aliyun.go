package sms

import (
	"encoding/json"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"xlangai/server/internal/repository"
)

func sendAliyun(cfg *repository.SmsServiceConfig, phone, code string) error {
	if cfg.APIKey == "" || cfg.SecretKey == "" {
		return fmt.Errorf("aliyun sms: missing api_key or secret_key")
	}
	if cfg.SignName == "" || cfg.TemplateCode == "" {
		return fmt.Errorf("aliyun sms: missing sign_name or template_code")
	}
	region := cfg.Region
	if region == "" {
		region = "cn-hangzhou"
	}
	endpoint := configString(cfg.Config, "endpoint", "dysmsapi.aliyuncs.com")
	paramKey := configString(cfg.Config, "template_param_key", "code")

	client, err := dysmsapi.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(cfg.APIKey),
		AccessKeySecret: tea.String(cfg.SecretKey),
		Endpoint:        tea.String(endpoint),
		RegionId:        tea.String(region),
	})
	if err != nil {
		return fmt.Errorf("aliyun sms client: %w", err)
	}

	params, err := json.Marshal(map[string]string{paramKey: code})
	if err != nil {
		return err
	}

	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(cfg.SignName),
		TemplateCode:  tea.String(cfg.TemplateCode),
		TemplateParam: tea.String(string(params)),
	}
	resp, err := client.SendSms(req)
	if err != nil {
		return fmt.Errorf("aliyun sms send: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("aliyun sms: empty response")
	}
	bodyCode := strings.TrimSpace(tea.StringValue(resp.Body.Code))
	if bodyCode != "" && bodyCode != "OK" {
		msg := tea.StringValue(resp.Body.Message)
		return fmt.Errorf("aliyun sms rejected: code=%s message=%s", bodyCode, msg)
	}
	return nil
}
