package auth

import "testing"

func TestGetToken_ManualToken(t *testing.T) {
	// 手动传入 token 应直接返回
	token, err := GetToken("my-jwt-token", "us")
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if token != "my-jwt-token" {
		t.Errorf("GetToken() = %q, want %q", token, "my-jwt-token")
	}
}

func TestGetToken_EmptyManualToken(t *testing.T) {
	// 空 token 应尝试调用 byte-auth
	// 在测试环境中 byte-auth 可能未安装，预期报错
	_, err := GetToken("", "us")
	if err == nil {
		// byte-auth 已安装且配置了 us 区域，这也是合法的
		return
	}
	// 应该包含有意义的错误信息
	if err.Error() == "" {
		t.Error("GetToken() should return meaningful error message")
	}
}

func TestIsByteAuthInstalled(t *testing.T) {
	// 仅测试函数不 panic
	_ = isByteAuthInstalled()
}
