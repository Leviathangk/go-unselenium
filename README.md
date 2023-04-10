# UnSelenium
实现类似 py 中 [undetected_chromedriver](https://github.com/ultrafunkamsterdam/undetected-chromedriver) 功能的 go 版本，可以过一些检测

该库参考了：[go-undetected_chromedriver](https://github.com/Davincible/go-undetected-chromedriver)

该库具备执行 CDP 命令的能力

不论是 py 还是 go，在这里的实现方式，本质上就是开启一个浏览器，通过 driver 去控制

注意：仅在 win 下测试通过，linux 未进行过测试

# 安装
```
go get -u github.com/Leviathangk/go-unselenium@latest
```

# 驱动
需要 [ChromeDriver](https://registry.npmmirror.com/binary.html?path=chromedriver/) 驱动，未安装请下载

# 使用
## Config
结构体如下： 
```
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
```

示例：创建 Config  
一般传 DriverPath 即可，如果是无头，设置下 Headless 就行
```
config := unselenium.NewConfig(
    unselenium.SetDriverPath("C:\\Program Files\\Python38\\chromedriver.exe"),
    unselenium.SetHeadless(),
    // unselenium.SetArgs("--headless=new"), // 自定义参数可以使用该方法
)
```
## Driver
结构体如下：
```
type Driver struct {
	Config             *Config            // 配置信息
	Chrome             *exec.Cmd          // 代表启动 Chrome 的进程
	Driver             *exec.Cmd          // 代表启动 Chrome 的进程
	ChromeCancel       context.CancelFunc // 取消执行函数
	DriverCancel       context.CancelFunc // 取消执行函数
	selenium.WebDriver                    // 代表 Selenium Driver：这样写就会拥有其方法
	HasStop            bool               // 是否已关闭
	Locker             sync.Mutex         // 锁：用来防止重复停止
}
```

示例：创建 Driver  
Driver 只要使用 selenium.WebDriver 下的方法就行，其余不用关注
```
driver, err := unselenium.NewDriver(unselenium.NewConfig(
    unselenium.SetDriverPath("C:\\Program Files\\Python38\\chromedriver.exe"),
    unselenium.SetHeadless(),
))
if err != nil {
    glog.Fatalln(err)
}
```

## 完整示例
```
func main() {
	// 启动 Driver
	driver, err := unselenium.NewDriver(unselenium.NewConfig(
		unselenium.SetDriverPath("C:\\Program Files\\Python38\\chromedriver.exe"),
		unselenium.SetHeadless(),
	))
	if err != nil {
		glog.Fatalln(err)
	}

	// 关闭浏览器及其服务（该方法被重写了）
	defer driver.Quit()

	// 测试 cdp 命令
	driver.ExecuteCDPScript("window.GK = 123;")

	// 检测点通过性查看
	// driver.Get("https://bot.sannysoft.com/")

	// 访问测试
	driver.Get("https://nowsecure.nl/")

	// 延迟关闭
	time.Sleep(10 * time.Second)

	// 保存图片测试，验证无头模式正常运行
	file, _ := os.Create("xxx.png")
	content, _ := driver.Screenshot()
	file.Write(content)
}
```