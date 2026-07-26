package domain

import (
	"testing"
)

func TestOtpMessageValidation(t *testing.T) {
	if _, err := NewOtpMessage("0550000101", "123456"); err != ErrInvalidPhone {
		t.Fatalf("local format = %v", err)
	}
	if _, err := NewOtpMessage("+233550000101", "12345"); err != ErrInvalidOtpCode {
		t.Fatalf("short code = %v", err)
	}
	message, err := NewOtpMessage("+233550000101", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if message.Template() != TemplateOtpCode || message.Params()["code"] != "123456" {
		t.Fatalf("message = %#v", message)
	}
}

func TestPodAlertValidation(t *testing.T) {
	if _, err := NewPodAlertMessage("+233550000101", " "); err != ErrPodRefRequired {
		t.Fatalf("blank ref = %v", err)
	}
	message, err := NewPodAlertMessage("+233550000101", "pod_1")
	if err != nil || message.Template() != TemplatePodAlert {
		t.Fatalf("message = %#v, %v", message, err)
	}
}
