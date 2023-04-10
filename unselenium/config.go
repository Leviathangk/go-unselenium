package unselenium

import (
	"errors"
	"strconv"
)

type Config struct {
	// 必传参数
	DriverPath string // 驱动路径

	// 无则自动设置
	ChromePath       string   // 浏览器路径
	ShowLog          bool     // 浏览器自己的日志：默认 false
	DriverArgs       []string // 启动 selenium 的配置
	UserDataDir      string   // 用户目录
	Language         string   // 语言：默认 zh-CN
	Headless         bool     // 无头模式：默认 false
	DisableMaxWindow bool     // 禁用全屏：默认 false
	Welcome          bool     // 启动欢迎页面：默认 false
	Sandbox          bool     // 沙箱：默认 false
	LogLevel         int      // 日志级别：默认 0

	// 自动配置：不需要自己设置
	Host       string // 远程连接的 ip
	ChromeAddr string // 远程连接的地址：chrome 的
	DriverAddr string // 远程连接的地址：driver 的
	ChromePort int    // 远程连接的端口：chrome 的
	DriverPort int    // 远程连接的端口：driver 的
}

type StartConfig func(c *Config)

// NewConfig 创建一个配置
func NewConfig(c ...StartConfig) *Config {
	config := new(Config)
	for _, sc := range c {
		sc(config)
	}

	return config
}

// Check 配置检查
func (config *Config) Check() error {
	// 必传参数检查
	if config.DriverPath == "" {
		return errors.New("Config 缺少 DriverPath 配置")
	}

	// 默认参数检查
	if config.Language == "" {
		SetLanguage("zh-CN")(config)
	}
	if !config.Welcome {
		config.setArgs("--no-default-browser-check", "--no-first-run")
	}
	if !config.Sandbox {
		config.setArgs("--no-sandbox", "--test-type")
	}
	if config.Headless {
		config.setArgs("--headless=new") // new 是支持 108 及以上版本的
	}
	if !config.DisableMaxWindow {
		config.setArgs("--window-size=1920,1080", "--start-maximized")
	}

	config.setArgs("--log-level=" + strconv.Itoa(config.LogLevel))

	return nil
}

// SetDriverPath 设置 Driver 路径
func SetDriverPath(path string) StartConfig {
	return func(c *Config) {
		c.DriverPath = path
	}
}

// SetShowLog 设置打印日志
func SetShowLog() StartConfig {
	return func(c *Config) {
		c.ShowLog = true
	}
}

// SetUserDataDir 设置用户目录
func SetUserDataDir(dir string) StartConfig {
	return func(c *Config) {
		c.UserDataDir = dir
	}
}

// SetHeadless 设置无头
func SetHeadless() StartConfig {
	return func(c *Config) {
		c.Headless = true
	}
}

// SetDisableMaxWindow 禁用全屏
func SetDisableMaxWindow() StartConfig {
	return func(c *Config) {
		c.DisableMaxWindow = true
	}
}

// SetWelcome 设置启动欢迎页面
func SetWelcome() StartConfig {
	return func(c *Config) {
		c.Welcome = true
	}
}

// SetSandbox 设置沙箱
func SetSandbox() StartConfig {
	return func(c *Config) {
		c.Sandbox = true
	}
}

// SetLogLevel 设置日志级别
func SetLogLevel(level int) StartConfig {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// SetLanguage 设置无头
func SetLanguage(language string) StartConfig {
	return func(c *Config) {
		c.Language = language
		c.DriverArgs = append(c.DriverArgs, "--lang="+language)
	}
}

// SetArgs 设置启动参数
func SetArgs(args ...string) StartConfig {
	return func(c *Config) {
		c.DriverArgs = append(c.DriverArgs, args...)
	}
}

// setArgs 设置启动参数
func (config *Config) setArgs(args ...string) {
	config.DriverArgs = append(config.DriverArgs, args...)
}
