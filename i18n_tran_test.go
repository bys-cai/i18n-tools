package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	language = "hi"
	filePath = "etc/hi.json"
)

func writeJSONFile(filename string, data interface{}) error {
	// 将数据序列化为JSON格式
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func TestParseJson(t *testing.T) {
	data, err := os.ReadFile("etc/locale/zh.json")
	newMap := make(map[string]interface{})

	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("解析JSON失败: %v", err)
	}
	for key, value := range result {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			childMap := make(map[string]interface{})
			// 是map类型，可以继续遍历
			for nestedKey, nestedValue := range nestedMap {
				s, err := translateValue(nestedValue)
				if err != nil {
					t.Fatalf("翻译失败: %v", err)
				}
				childMap[nestedKey] = s
				time.Sleep(2 * time.Second)
			}
			newMap[key] = childMap
		} else {
			s, err := translateValue(value)

			if err != nil {
				t.Fatalf("翻译失败: %v", err)
			}
			newMap[key] = s
			time.Sleep(1 * time.Second)
			// 不是map类型，直接使用值
		}
	}
	err = writeJSONFile(filePath, newMap)
	if err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

}
func translateValue(value interface{}) (string, error) {
	var msg string
	switch v := value.(type) {
	case string:
		msg = v
	case []byte:
		// 处理字节切片，确保UTF-8编码
		if utf8.Valid(v) {
			msg = string(v)
		} else {
			msg = fmt.Sprintf("%s", value)
		}
	default:
		msg = fmt.Sprintf("%s", value)
	}

	// 确保消息是有效的UTF-8字符串
	if !utf8.ValidString(msg) {
		return "", fmt.Errorf("无效的UTF-8字符串")
	}

	return tranApi(msg, language)
}

func tranApi(msg string, to string) (string, error) {
	// 配置参数 - 建议使用环境变量或配置文件
	const (
		http_url = "https://fanyi-api.baidu.com/api/trans/vip/translate" // 替换为实际的翻译API地址
		apiKey   = "omavlKoUrFVyJ8OekLt8"                                // 替换为控制台生成的API key
		appID    = "20230320001606902"
		salt     = "65478" // 替换为控制台生成的appid
	)
	sign := generateSign(appID, msg, salt, apiKey)
	params := url.Values{}

	params.Set("sign", sign)
	params.Set("salt", salt)
	// 准备请求数据
	params.Set("appid", appID)
	params.Set("from", "zh")
	params.Set("to", to)
	params.Set("q", fmt.Sprintf("%s", msg))

	// 设置请求头
	fullURL := http_url + "?" + params.Encode()
	// 创建并发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取并解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}
	fmt.Println("响应" + string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析JSON失败: %v, 原始响应: %s", err, string(body))
	}

	// 验证响应结构
	if errMsg, exists := result["error"]; exists {
		return "", fmt.Errorf("API返回错误: %v", errMsg)
	}

	// 如果有翻译结果，可以进行断言测试
	// 在您的测试函数中添加解析逻辑
	if transResult, exists := result["trans_result"]; exists {
		logx.Infof("翻译结果: %+v", transResult)
		// 类型断言转换为切片
		if results, ok := transResult.([]interface{}); ok && len(results) > 0 {
			// 获取第一个翻译结果
			if firstResult, ok := results[0].(map[string]interface{}); ok {
				dstText := firstResult["dst"].(string)
				return dstText, nil
			}
		}
	}

	return "", nil

}

func TestTranBaidu(t *testing.T) {
	api, err := tranApi("你好", "en")
	if err != nil {
		t.Fatalf("翻译失败: %v", err)
	}
	t.Logf("翻译结果: %s", api)
}

func generateSign(appID, query, salt, appKey string) string {
	// 构造签名字符串: appid + q + salt + appkey
	signStr := appID + query + salt + appKey

	// 对字符串进行MD5加密
	hasher := md5.New()
	hasher.Write([]byte(signStr))

	// 返回32位小写MD5值
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
