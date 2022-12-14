package clock

import (
	"container/list"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type CallBack func(any)

type Clock struct {
	interval      time.Duration
	ticker        *time.Ticker
	slots         []*list.List
	slotNum       int
	callback      func()
	timer         map[string]int
	currentPos    int
	addTaskSig    chan Task
	removeTaskSig chan string
	stopSig       chan bool
	lock          *sync.Mutex
	deleteList    []string
}

type Task struct {
	delay    time.Duration // 延迟时间
	circle   int           // 时间轮需要转动几圈
	key      string        // 定时器唯一标识, 用于删除定时器
	callback CallBack
	data     any // 回调函数参数
}

// New 创建时间轮
func New(interval time.Duration, slotNum int, cb func()) *Clock {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}
	clock := &Clock{
		interval:      interval,
		slots:         make([]*list.List, slotNum),
		timer:         make(map[string]int),
		currentPos:    0,
		callback:      cb,
		slotNum:       slotNum,
		addTaskSig:    make(chan Task),
		removeTaskSig: make(chan string),
		stopSig:       make(chan bool),
		lock:          &sync.Mutex{},
		deleteList:    make([]string, 1024),
	}

	clock.initSlots()

	return clock
}

// 初始化槽，每个槽指向一个双向链表
func (c *Clock) initSlots() {
	for i := 0; i < c.slotNum; i++ {
		c.slots[i] = list.New()
	}
}

// Start 启动时间轮
func (c *Clock) Start() {
	if !c.lock.TryLock() {
		log.Println("already started")
		return
	}

	c.ticker = time.NewTicker(c.interval)
	c.lock.Lock()
	go c.start()

}

// Stop 停止时间轮
func (c *Clock) Stop() {
	defer c.callback()
	c.stopSig <- true
}

// AddTimer 添加定时器 key为定时器唯一标识
func (c *Clock) AddTimer(delay time.Duration, cb CallBack, data any) *string {
	if delay < 0 {
		return nil
	}
	id := uuid.NewString()
	c.addTaskSig <- Task{delay: delay, key: id, callback: cb, data: data}
	log.Printf("[+] New Task: %s\n", id)
	return &id
}

// RemoveTimer 删除定时器 key为添加定时器时传递的定时器唯一标识
// TODO:增加判斷允許key使用 uuid string 或者 uint32
func (c *Clock) RemoveTimer(key string) {
	c.removeTaskSig <- key

}

func (c *Clock) start() {
	for {
		select {
		case <-c.ticker.C:
			c.tickHandler()
		case task := <-c.addTaskSig:
			c.addTask(&task)
		case key := <-c.removeTaskSig:
			c.removeTask(key)
		case <-c.stopSig:
			c.ticker.Stop()
			c.lock.Unlock()
			return
		}
	}
}

func (c *Clock) tickHandler() {
	l := c.slots[c.currentPos]
	c.scanAndRunTask(l)
	if c.currentPos == c.slotNum-1 {
		c.currentPos = 0
	} else {
		c.currentPos++
	}
}

// 扫描链表中过期定时器, 并执行回调函数
func (c *Clock) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		for _, v := range c.deleteList {
			if task.key == v {
				delete(c.timer, task.key)
				l.Remove(e)
			}
		}
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}

		//一次性 Task
		go task.callback(task.data)
		next := e.Next()
		l.Remove(e)
		delete(c.timer, task.key)

		e = next
	}
}

// 新增任务到链表中
func (c *Clock) addTask(task *Task) {
	pos, circle := c.getPositionAndCircle(task.delay)
	task.circle = circle

	c.slots[pos].PushBack(task)

	c.timer[task.key] = pos
}

// 获取定时器在槽中的位置, 时间轮需要转动的圈数
func (c *Clock) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(c.interval.Seconds())
	circle = delaySeconds / intervalSeconds / c.slotNum
	pos = (c.currentPos + delaySeconds/intervalSeconds) % c.slotNum

	return
}

// 从链表中删除任务
func (c *Clock) removeTask(key string) {
	c.deleteList = append(c.deleteList, key)
}
