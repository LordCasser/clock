# clock

![img](https://goreportcard.com/badge/github.com/lordcasser/clock)

forked from https://github.com/ouqiang/timewheel

Golang实现的时间轮



![时间轮](https://raw.githubusercontent.com/ouqiang/timewheel/master/timewheel.jpg)


## 安装

```shell
go get -u github.com/LordCasser/clock
```

## 使用

```go
package main
import (
    "github.com/LordCasser/clock"
    "time"
)

func main() {
    // 第一个参数为tick刻度, 即时间轮多久转动一次
    // 第二个参数为时间轮槽slot数量
    // 第三个参数为回调函数,在整个时间轮退出时调用
	c := New(1*time.Second, 10, func() {
		log.Println("[*] Finish")
	})

	c.Start()
	
	var list []string
	var data = "gogogo"

    // 添加定时器 
    // 第一个参数为延迟时间
    // 第二个参数为定时器触发时调用的函数
    // 第三个参数为用户自定义数据, 此参数将会传递给回调函数, 类型为interface{}
    // 返回值是uuid格式的唯一ID，类型为string
	id := c.AddTimer(5*time.Second, func(data any) {
		log.Printf("test: %s", Strval(data))
	}, data)
	list = append(list, *id)


	id2 := c.AddTimer(11*time.Second, func(data any) {
		log.Printf("test2: %s", Strval(data))
	}, data)
	list = append(list, *id2)
	select {}
}
```

