package unselenium

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
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

var (
	signalChannel = make(chan os.Signal, 3) // 信号监听通道
	Drivers       []*Driver                 // 所有已经启动的 driver
	ExitWhenKill  = true                    // 在监听到退出信号时，主动触发 os.Exit()
)

func init() {
	// 监听指定信号
	// os.Interrupt 为 ctrl+c
	// os.Kill 为 kill
	// syscall.SIGTERM 为 kill 不加 -9 时的 pid
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		// 阻塞，直到有信号传入
		s := <-signalChannel
		glog.Debugf("An exit signal is received：%s\n", s)

		StopAll() // 释放所有资源

		// 退出系统
		if ExitWhenKill {
			os.Exit(1)
		}
	}()
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

	// 存放已启动的 Driver，保证资源释放
	Drivers = append(Drivers, driver)

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

		// 移除 driver
		for index, driver := range Drivers {
			if d == driver {
				Drivers = append(Drivers[:index], Drivers[index+1:]...)
				break
			}
		}

		d.HasStop = true
		glog.Debugln("UnSelenium Closed Successfully!")
	}
}

// StopAll 释放所有已启动的 Driver 的资源
func StopAll() {
	for _, driver := range Drivers {
		driver.Quit()
	}
}
