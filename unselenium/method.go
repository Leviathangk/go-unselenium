package unselenium

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/Leviathangk/go-glog/glog"
	"github.com/stitch-june/selenium"
	"github.com/stitch-june/selenium/chrome"
)

// getAddress 获取调试的端口
func (config *Config) getAddress() (string, int, error) {
	var host string
	if config.Host == "" {
		host = "127.0.0.1"
		config.Host = host
	} else {
		host = config.Host
	}
	port := "0"

	addr := host + ":" + port
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return "", 0, fmt.Errorf("获取启动端口失败：%s，引发错误：%v", addr, err)
	}
	defer l.Close()

	split := strings.Split(l.Addr().String(), ":")

	if len(split) > 1 {
		port = split[1]
	} else {
		port = split[0]
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, fmt.Errorf("获取启动端口失败：%s，引发错误：%v", addr, err)
	}

	return host, portInt, nil
}

// setAddress 设置调试的端口
func (config *Config) setAddress() error {
	host, port, err := config.getAddress()
	if err != nil {
		return err
	}

	// 添加远程连接配置
	portStr := strconv.Itoa(port)
	config.setArgs("--remote-debugging-host=" + host)
	config.setArgs("--remote-debugging-port=" + portStr)

	// 修改设置
	config.Host = host
	config.ChromePort = port
	config.ChromeAddr = host + ":" + portStr

	return nil
}

// setUserData 设置用户目录
func (config *Config) setUserData() error {
	if config.UserDataDir != "" {
		config.setArgs("--user-data-dir=" + config.UserDataDir)
		return nil
	}

	dir, err := os.MkdirTemp("", "undetected-chromedriver-userdata-*")
	if err != nil {
		return fmt.Errorf("设置用户文件夹失败：%s", dir)
	}

	config.UserDataDir = dir
	config.setArgs("--user-data-dir=" + dir)

	return nil
}

// startChrome 打开一个浏览器
func (d *Driver) startChrome() error {
	if d.Config.ChromePath == "" {
		d.Config.ChromePath = findChromePath()
	}

	if d.Config.ChromePath == "" {
		return fmt.Errorf("未提取到 ChromePath，请自行输入")
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.ChromeCancel = cancel
	d.Chrome = exec.CommandContext(ctx, d.Config.ChromePath, d.Config.DriverArgs...)

	if d.Config.ShowLog {
		d.Chrome.Stdout = os.Stdout
		d.Chrome.Stderr = os.Stderr
	}

	glog.Debugf("Starting Chrome Cmd: %s\n", d.Chrome.String())

	if err := d.Chrome.Start(); err != nil {
		return fmt.Errorf("failed to start chrome: %w", err)
	}

	return nil
}

// startDriver 启动 driver
func (d *Driver) startDriver() error {
	host, port, err := d.Config.getAddress()
	if err != nil {
		return nil
	}
	d.Config.DriverPort = port
	d.Config.DriverAddr = host + ":" + strconv.Itoa(port)

	ctx, cancel := context.WithCancel(context.Background())
	d.DriverCancel = cancel
	d.Driver = exec.CommandContext(ctx, d.Config.DriverPath, "--port="+strconv.Itoa(port))

	if d.Config.ShowLog {
		d.Driver.Stdout = os.Stdout
		d.Driver.Stderr = os.Stderr
	}

	glog.Debugf("Starting ChromeDriver Cmd: %s\n", d.Driver.String())

	if err := d.Driver.Start(); err != nil {
		return fmt.Errorf("failed to start chromedriver: %w", err)
	}

	return nil
}

// connect 与浏览器连接
func (d *Driver) connect() error {
	caps := selenium.Capabilities{
		"browserName":      "chrome",
		"pageLoadStrategy": "normal",
	}

	caps.AddChrome(chrome.Capabilities{
		Path:         d.Config.ChromePath,
		Args:         d.Config.DriverArgs,
		DebuggerAddr: d.Config.ChromeAddr,
	})

	addr := fmt.Sprintf("http://%s", d.Config.DriverAddr)

	glog.Debugf("Connecting Addr: %s\n", addr)

	driver, err := selenium.NewRemote(caps, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to chromedriver: %w", err)
	}

	d.WebDriver = driver

	return nil
}

// findChromePath 寻找浏览器的路径
func findChromePath() string {
	goos := strings.ToLower(runtime.GOOS)

	switch {
	case goos == "linux" || goos == "darwin":
		binaries := []string{
			"google-chrome",
			"chromium",
			"chromium-browser",
			"chrome",
			"google-chrome-stable",
		}

		for _, bin := range binaries {
			if p, err := exec.LookPath(bin); err == nil {
				return p
			}
		}
	case goos == "windows":
		paths := []string{
			"PROGRAMFILES",
			"PROGRAMFILES(X86)",
			"LOCALAPPDATA",
			"PROGRAMW6432",
		}
		subitems := []string{
			"Google/Chrome/Application/chrome.exe",
			"Google/Chrome Beta/Application/chrome.exe",
			"Google/Chrome Canary/Application/chrome.exe",
		}

		for _, p := range paths {
			e := os.Getenv(p)
			if e != "" {
				for _, subitem := range subitems {
					if pp, err := exec.LookPath(strings.ReplaceAll(e, "\\", "/") + "/" + subitem); err == nil {
						return pp
					}
				}
			}
		}
	}

	return ""
}

func (d *Driver) getCdcProps() bool {
	script := `
    let objectToInspect = window;
    let result = [];
    while (objectToInspect !== null) {
        result = result.concat(Object.getOwnPropertyNames(objectToInspect));
        objectToInspect = Object.getPrototypeOf(objectToInspect);
    }
	return result.filter(i=>i.match(/.+_.+_(Array|Promise|Symbol)/ig))
	`

	res, err := d.ExecuteScript(script, nil)
	if err != nil {
		glog.Warningln("Failed to execute get cdc script", err)
		return false
	}

	return len(res.([]any)) > 0
}

func (d *Driver) removeCdcProps() {
	script := `
	let objectToInspect = window;
	let result = [];
	while (objectToInspect !== null) {
		result = result.concat(Object.getOwnPropertyNames(objectToInspect));
		objectToInspect = Object.getPrototypeOf(objectToInspect);
	}
	result.forEach(p=>p.match(/.+_.+_(Array|Promise|Symbol)/ig) && delete window[p])
	`
	if _, err := d.ExecuteCDPScript(script); err != nil {
		glog.Warningln("Execute remove cdc props", err)
	}
}

// monitorExit 监控退出信号：保证资源释放
func (d *Driver) monitorExit() {
	// 创建通道
	signalChannel := make(chan os.Signal)

	// 监听指定信号
	// os.Interrupt 为 ctrl+c
	// os.Kill 为 kill
	// syscall.SIGTERM 为 kill 不加 -9 时的 pid
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		s := <-signalChannel // 阻塞，直到有信号传入
		glog.Debugf("An exit signal is received：%s\n", s)
		d.Quit()
	}()
}

// ExecuteCDP 执行 CDP 命令
func (d *Driver) ExecuteCDP(cmd string, params interface{}) (interface{}, error) {
	return d.ExecuteChromeDPCommand(cmd, params)
}

// ExecuteCDPScript 打开新页面时执行脚本
func (d *Driver) ExecuteCDPScript(script string) (interface{}, error) {
	return d.ExecuteChromeDPCommand("Page.addScriptToEvaluateOnNewDocument", map[string]string{"source": script})
}
