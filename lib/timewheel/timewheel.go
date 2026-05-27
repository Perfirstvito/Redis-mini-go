package timewheel

// 时间轮算法，高校调度延迟任务
// key 过期删除，事务超时回滚，分布式锁过期删除等场景都可以使用时间轮算法来实现

import (
	"container/list"
	"my-redis/lib/logger"
	"sync"
	"time"
)

type location struct {
	slot  int           // 所在槽索引
	etask *list.Element // 链表元素指针（用于 O(1) 删除）
}

// TimeWheel can execute jobs after a given delay
type TimeWheel struct {
	interval time.Duration // 每格的时间粒度（如 1 秒）
	ticker   *time.Ticker  // 定时器，驱动时间轮转动
	slots    []*list.List  // 时间格数组，每格一个双向链表

	timer      map[string]*location // key → 位置（支持 O(1) 定位）
	currentPos int                  // 当前指向的槽位置
	slotNum    int                  // 槽数量

	addTaskChannel    chan task   // 添加任务通道（线程安全）
	removeTaskChannel chan string // 删除任务通道
	stopChannel       chan bool   // 停止信号通道

	mu sync.RWMutex // 保护链表扫描
}

type task struct {
	delay  time.Duration // 延迟时间
	circle int           // 需要转完整圈的次数
	key    string        // 任务唯一标识（用于取消）
	job    func()        // 实际执行函数
}

// New creates a new time wheel
func New(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[string]*location),
		currentPos:        0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan task),
		removeTaskChannel: make(chan string),
		stopChannel:       make(chan bool),
	}
	tw.initSlots()
	return tw
}

func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

// Start starts the time wheel
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

// Stop stops the time wheel
func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

// AddJob adds a new job to the pending queue
func (tw *TimeWheel) AddJob(delay time.Duration, key string, job func()) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- task{delay: delay, key: key, job: job}
}

// RemoveJob add remove job from pending queue
// if job is done or not found, then nothing happened
func (tw *TimeWheel) RemoveJob(key string) {
	if key == "" {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	tw.mu.Lock()
	l := tw.slots[tw.currentPos]
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
	tw.mu.Unlock()

	go tw.scanAndRunTask(l)
}

func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	var tasksToRemove []string
	tw.mu.RLock() // Read lock for accessing the list
	for e := l.Front(); e != nil; {
		task := e.Value.(*task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}

		go func(job func()) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(err)
				}
			}()
			job()
		}(task.job)

		if task.key != "" {
			tasksToRemove = append(tasksToRemove, task.key)
		}
		next := e.Next()
		l.Remove(e) // Safe as this is a local operation
		e = next
	}
	tw.mu.RUnlock()

	// Remove tasks from the timer after the scan
	tw.mu.Lock()
	for _, key := range tasksToRemove {
		delete(tw.timer, key)
	}
	tw.mu.Unlock()
}

func (tw *TimeWheel) addTask(task *task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	tw.mu.Lock()
	defer tw.mu.Unlock()

	if task.key != "" {
		if _, ok := tw.timer[task.key]; ok {
			tw.removeTaskInternal(task.key) // Internal version avoids double lock
		}
	}

	e := tw.slots[pos].PushBack(task)
	loc := &location{
		slot:  pos,
		etask: e,
	}
	tw.timer[task.key] = loc
}

func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	circle = delaySeconds / intervalSeconds / tw.slotNum
	pos = (tw.currentPos + delaySeconds/intervalSeconds) % tw.slotNum
	return
}

func (tw *TimeWheel) removeTask(key string) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.removeTaskInternal(key)
}

func (tw *TimeWheel) removeTaskInternal(key string) {
	pos, ok := tw.timer[key]
	if !ok {
		return
	}
	l := tw.slots[pos.slot]
	l.Remove(pos.etask)
	delete(tw.timer, key)
}
