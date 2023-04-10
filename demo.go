package main

import (
	"github.com/Leviathangk/go-unselenium/unselenium"
	"time"

	"github.com/Leviathangk/go-glog/glog"
)

func main() {
	// 启动 Driver
	driver, err := unselenium.NewDriver(unselenium.NewConfig(
		unselenium.SetDriverPath("C:\\Program Files\\Python39\\chromedriver.exe"),
		//unselenium.SetHeadless(),
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
	//file, _ := os.Create("xxx.png")
	//content, _ := driver.Screenshot()
	//file.Write(content)
}
