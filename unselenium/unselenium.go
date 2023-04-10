package unselenium

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/Leviathangk/go-glog/glog"
	"github.com/stitch-june/selenium"
)

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

// NewDriver 创建 Driver
func NewDriver(config *Config) (*Driver, error) {
	driver := new(Driver)
	driver.Config = config

	// 配置检查
	err := config.Check()
	if err != nil {
		return nil, err
	}

	// 设置远程连接地址
	err = config.setAddress()
	if err != nil {
		return nil, err
	}

	// 设置用户目录
	err = config.setUserData()
	if err != nil {
		return nil, err
	}

	// 启动浏览器
	err = driver.startChrome()
	if err != nil {
		driver.Quit()
		return nil, err
	}

	// 启动 Driver
	err = driver.startDriver()
	if err != nil {
		driver.Quit()
		return nil, err
	}

	time.Sleep(3 * time.Second)

	// Driver 与 Chrome 建立连接
	err = driver.connect()
	if err != nil {
		driver.Quit()
		return nil, err
	}

	glog.Debugln("UnSelenium Start Successfully!")

	return driver, nil
}

// Get 发出请求
func (d *Driver) Get(url string) error {
	if d.getCdcProps() {
		d.removeCdcProps()
	}

	return d.WebDriver.Get(url)
}

// Quit 终结执行
func (d *Driver) Quit() {
	if !d.HasStop && d.Locker.TryLock() {
		defer d.Locker.Unlock()
		glog.Debugln("Exiting UnSelenium...")
		if d.ChromeCancel != nil {
			d.ChromeCancel()
		}
		if d.DriverCancel != nil {
			d.DriverCancel()
		}
		if d.WebDriver != nil {
			d.WebDriver.Quit()
		}
		d.HasStop = true
		glog.Debugln("UnSelenium Closed Successfully!")
	}
}
