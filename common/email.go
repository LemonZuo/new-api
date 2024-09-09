package common

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

func generateMessageID() string {
	// 生成时间戳和随机字符串
	timestamp := time.Now().UnixNano()
	randomStr := GetRandomString(12)

	// 使用域名或服务器地址初始化变量
	domainOrServer := SMTPServer

	// 如果 SMTPAccount 包含 '@'，提取域名
	if atPos := strings.Index(SMTPAccount, "@"); atPos != -1 {
		if atPos+1 < len(SMTPAccount) {
			domainOrServer = SMTPAccount[atPos+1:]
		}
	}

	// 生成并返回消息ID
	return fmt.Sprintf("<%d.%s@%s>", timestamp, randomStr, domainOrServer)
}

func SendEmail(subject string, receiver string, content string) error {
	if SMTPFrom == "" { // for compatibility
		SMTPFrom = SMTPAccount
	}
	if SMTPServer == "" && SMTPAccount == "" {
		return fmt.Errorf("SMTP 服务器未配置")
	}
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
	mail := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s<%s>\r\n"+
		"Subject: %s\r\n"+
		"Date: %s\r\n"+
		"Message-ID: %s\r\n"+ // 添加 Message-ID 头
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, SystemName, SMTPFrom, encodedSubject, time.Now().Format(time.RFC1123Z), generateMessageID(), content))
	auth := smtp.PlainAuth("", SMTPAccount, SMTPToken, SMTPServer)
	addr := fmt.Sprintf("%s:%d", SMTPServer, SMTPPort)
	to := strings.Split(receiver, ";")
	var err error
	if SMTPPort == 465 || SMTPSSLEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         SMTPServer,
		}
		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", SMTPServer, SMTPPort), tlsConfig)
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, SMTPServer)
		if err != nil {
			return err
		}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(SMTPFrom); err != nil {
			return err
		}
		receiverEmails := strings.Split(receiver, ";")
		for _, receiver := range receiverEmails {
			if err = client.Rcpt(receiver); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(mail)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
	} else if isOutlookServer(SMTPAccount) {
		auth = LoginAuth(SMTPAccount, SMTPToken)
		err = smtp.SendMail(addr, auth, SMTPAccount, to, mail)
	} else {
		err = smtp.SendMail(addr, auth, SMTPAccount, to, mail)
	}
	return err
}
