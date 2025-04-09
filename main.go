package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// headers 是一个自定义的类型，用于存储多个 Header
type headers map[string]string

// String 方法用于实现 flag.Value 接口中的 String 方法
func (h *headers) String() string {
	var result []string
	for key, val := range *h {
		result = append(result, fmt.Sprintf("%s: %s", key, val))
	}
	return strings.Join(result, ", ")
}

// Set 方法用于实现 flag.Value 接口中的 Set 方法，用于解析并存储 Header
func (h *headers) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format, should be key:value")
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	(*h)[key] = val
	return nil
}

func main() {
	var customHeaders headers = make(map[string]string)
	// 定义一个 -H 参数，用于设置自定义 Header
	flag.Var(&customHeaders, "H", "HTTP header in key:value format")
	env := flag.String("h", "", "Host eg: https://www.baidu.com embed prod or dev or local")
	path := flag.String("p", "", "Request path")
	method := flag.String("X", "GET", "HTTP method")
	port := flag.Int("P", 59447, "Port")
	body := flag.String("d", "", "Request body (auto convert JSON to form data)")
	jsonBody := flag.String("json", "", "JSON request body (with Content-Type: application/json)")
	token := flag.String("t", "", "Authorization token")

	flag.Parse()
	// 如果用户没有输入参数，则打印帮助信息
	if *env == "" {
		flag.Usage()
		os.Exit(0)
	}

	if *token != "" {
		customHeaders["Authorization"] = *token
	}

	requestUrl := getRequestUrl(*env, *port, *path)
	
	// 选择使用哪种请求体和对应的 Content-Type
	var requestBody string
	var contentType string
	
	if *jsonBody != "" {
		requestBody = *jsonBody
		contentType = "application/json"
	} else if *body != "" {
			var jsonMap map[string]interface{}
			err := json.Unmarshal([]byte(*body), &jsonMap)
			if err == nil {
				// 成功解析为 JSON，转换为表单格式
				params := url.Values{}
				convertJSONToForm(jsonMap, "", params)
				requestBody = params.Encode()
			} else {
				// 解析失败，保持原样
				fmt.Printf("警告: JSON 解析失败: %v\n", err)
				requestBody = *body
			}
		contentType = "application/x-www-form-urlencoded"
	} 	
	customHeaders["Content-Type"] = contentType
	
	doRequest(*method, requestUrl, requestBody, &customHeaders)
}

func doRequest(method string, url string, reqBody string, h *headers) {
	req, err := http.NewRequest(method, url, strings.NewReader(reqBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	for key, val := range *h {
		req.Header.Add(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response from %s:\n%s\n", url, string(body))
}

func getRequestUrl(env string, port int, path string) string {
	url := ""
	switch env {
	case "prod":
		url = "https://www.prod.com" // 替换为你的生产环境 url
	case "dev":
		url = "https://www.dev.com" // 替换为你的开发环境 url
	case "local":
		url = "http://localhost:" + fmt.Sprint(port) // 替换为你的本地环境 url
	default:
		url = env
	}
	url = url + path
	return url
}

// convertJSONToForm 将 JSON 对象递归转换为表单数据
// parentKey 是父级键名，用于处理嵌套对象
// params 是表单数据的结果集
func convertJSONToForm(jsonMap map[string]interface{}, parentKey string, params url.Values) {
	for key, value := range jsonMap {
		// 构建当前键名
		var currentKey string
		if parentKey == "" {
			currentKey = key
		} else {
			// 对于嵌套对象，使用 parent[child] 格式
			currentKey = fmt.Sprintf("%s[%s]", parentKey, key)
		}
		
		// 根据值的类型进行处理
		switch v := value.(type) {
		case map[string]interface{}:
			// 递归处理嵌套对象
			convertJSONToForm(v, currentKey, params)
		case []interface{}:
			// 处理数组
			for i, item := range v {
				arrayKey := fmt.Sprintf("%s[%d]", currentKey, i)
				// 如果数组元素是对象，递归处理
				if mapItem, isMap := item.(map[string]interface{}); isMap {
					convertJSONToForm(mapItem, arrayKey, params)
				} else {
					// 简单类型直接添加
					params.Add(arrayKey, fmt.Sprintf("%v", item))
				}
			}
		default:
			// 简单类型直接添加
			params.Add(currentKey, fmt.Sprintf("%v", value))
		}
	}
}
