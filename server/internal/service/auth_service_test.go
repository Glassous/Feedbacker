package service

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	const length = 16

	first, err := GeneratePassword(length)
	if err != nil {
		t.Fatalf("生成密码失败: %v", err)
	}
	if len(first) != length {
		t.Fatalf("密码长度 = %d, want %d", len(first), length)
	}

	for _, char := range first {
		if !strings.ContainsRune(passwordAlphabet, char) {
			t.Fatalf("密码包含非法字符 %q", char)
		}
	}

	second, err := GeneratePassword(length)
	if err != nil {
		t.Fatalf("生成密码失败: %v", err)
	}
	if first == second {
		t.Fatal("两次生成的密码不应相同")
	}
}

func TestPasswordAlphabetExcludesAmbiguousChars(t *testing.T) {
	for _, char := range "0O1lI" {
		if strings.ContainsRune(passwordAlphabet, char) {
			t.Fatalf("密码字符集不应包含易混淆字符 %q", char)
		}
	}
}
