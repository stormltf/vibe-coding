package i18n

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// 语言代码常量
const (
	ZhCN = "zh-CN" // 简体中文
	EnUS = "en-US" // 英语
)

// DefaultLang 默认语言
var DefaultLang = ZhCN

// langKey context 中存储语言的 key
type langKey struct{}

// Message 消息定义
type Message map[string]string

// Bundle 语言包
type Bundle struct {
	messages map[string]Message // map[lang]map[key]message
	mu       sync.RWMutex
}

// 全局语言包
var bundle = &Bundle{
	messages: make(map[string]Message),
}

// LoadFromFS 从嵌入文件系统加载翻译
func LoadFromFS(fs embed.FS, pattern string) error {
	files, err := fs.ReadDir(".")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		data, err := fs.ReadFile(name)
		if err != nil {
			return err
		}

		// 从文件名提取语言代码 (例如: zh-CN.yaml -> zh-CN)
		lang := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		if err := LoadFromBytes(lang, data); err != nil {
			return err
		}
	}

	return nil
}

// LoadFromBytes 从字节数据加载翻译
func LoadFromBytes(lang string, data []byte) error {
	var messages Message
	if err := yaml.Unmarshal(data, &messages); err != nil {
		return err
	}

	bundle.mu.Lock()
	bundle.messages[lang] = messages
	bundle.mu.Unlock()

	return nil
}

// LoadMessages 直接加载消息
func LoadMessages(lang string, messages Message) {
	bundle.mu.Lock()
	bundle.messages[lang] = messages
	bundle.mu.Unlock()
}

// T 翻译消息
func T(lang, key string, args ...interface{}) string {
	bundle.mu.RLock()
	defer bundle.mu.RUnlock()

	// 先尝试指定语言
	if msgs, ok := bundle.messages[lang]; ok {
		if msg, ok := msgs[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// 回退到默认语言
	if msgs, ok := bundle.messages[DefaultLang]; ok {
		if msg, ok := msgs[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// 返回 key 本身
	return key
}

// Tr 从 context 获取语言并翻译
func Tr(ctx context.Context, key string, args ...interface{}) string {
	lang := GetLang(ctx)
	return T(lang, key, args...)
}

// WithLang 将语言设置到 context
func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, langKey{}, lang)
}

// GetLang 从 context 获取语言
func GetLang(ctx context.Context) string {
	if lang, ok := ctx.Value(langKey{}).(string); ok {
		return lang
	}
	return DefaultLang
}

// ParseAcceptLanguage 解析 Accept-Language 请求头
func ParseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return DefaultLang
	}

	// 简单解析，取第一个语言
	parts := strings.Split(acceptLang, ",")
	if len(parts) == 0 {
		return DefaultLang
	}

	lang := strings.TrimSpace(strings.Split(parts[0], ";")[0])

	// 标准化语言代码
	switch strings.ToLower(lang) {
	case "zh", "zh-cn", "zh-hans":
		return ZhCN
	case "en", "en-us":
		return EnUS
	default:
		// 尝试直接使用
		bundle.mu.RLock()
		_, ok := bundle.messages[lang]
		bundle.mu.RUnlock()
		if ok {
			return lang
		}
		return DefaultLang
	}
}

// SupportedLanguages 返回支持的语言列表
func SupportedLanguages() []string {
	bundle.mu.RLock()
	defer bundle.mu.RUnlock()

	langs := make([]string, 0, len(bundle.messages))
	for lang := range bundle.messages {
		langs = append(langs, lang)
	}
	return langs
}
