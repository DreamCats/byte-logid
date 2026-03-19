// Package auth 提供 JWT 认证功能，支持委托 byte-auth CLI 或手动传入 token。
package auth

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetToken 获取 JWT Token。
// 优先使用手动传入的 token（--token 参数），否则调用 byte-auth CLI 获取。
//
// 参数:
//   - manualToken: 用户通过 --token 传入的 token，为空则自动获取
//   - region: 目标区域标识（us/i18n/eu/cn）
//
// 返回 JWT token 字符串，出错时返回 error。
func GetToken(manualToken string, region string) (string, error) {
	if manualToken != "" {
		return manualToken, nil
	}
	return callByteAuth(region)
}

// callByteAuth 调用 byte-auth CLI 获取 JWT token。
// 执行命令: byte-auth token --region <region> --raw
func callByteAuth(region string) (string, error) {
	if !isByteAuthInstalled() {
		return "", fmt.Errorf("byte-auth 未安装，请先安装: go install github.com/DreamCats/byte-auth@latest")
	}

	cmd := exec.Command("byte-auth", "token", "--region", region, "--raw")
	output, err := cmd.Output()
	if err != nil {
		// 尝试获取 stderr 的错误信息
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			return "", fmt.Errorf("byte-auth 执行失败: %s\n请检查 byte-auth 配置: byte-auth config show --region %s", stderr, region)
		}
		return "", fmt.Errorf("byte-auth 执行失败: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("byte-auth 返回了空 token，请执行 byte-auth token --region %s --force 刷新", region)
	}

	return token, nil
}

// isByteAuthInstalled 检查 byte-auth 是否已安装在 PATH 中。
func isByteAuthInstalled() bool {
	_, err := exec.LookPath("byte-auth")
	return err == nil
}
