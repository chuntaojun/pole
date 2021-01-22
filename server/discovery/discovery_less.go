package discovery

import (
	"sync"
	"time"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/utils"
)

type ExpireLessWatcher func(keys []string)

type LessHolder struct {
	ctx        *common.ContextPole
	lessLock   sync.RWMutex
	lessRep    map[string]time.Duration
	rwLock     sync.RWMutex
	expireKeys chan []string
	watchers   []ExpireLessWatcher
	timer      *utils.HashTimeWheel
}

func NewLessHolder(ctx *common.ContextPole) *LessHolder {
	lh := &LessHolder{
		ctx:      ctx,
		lessLock: sync.RWMutex{},
		lessRep:  make(map[string]time.Duration),
		rwLock:   sync.RWMutex{},
		watchers: make([]ExpireLessWatcher, 0, 0),
		timer:    utils.NewTimeWheel(time.Duration(1)*time.Second, 15),
	}

	lh.start()
	return lh
}

func (lh *LessHolder) start() {
	utils.Go(lh.ctx.NewSubCtx(), func(ctx *common.ContextPole) {
		for {
			select {
			case <-ctx.Done():
				return
			case keys := <-lh.expireKeys:
				lh.notifyExpire(keys)
			}
		}
	})
}

func (lh *LessHolder) GrantLess(instance Instance) {
	key := instance.GetKey()
	defer lh.lessLock.Unlock()
	lh.lessLock.Lock()
	expire := time.Duration(15+time.Now().Unix()) * time.Second
	lh.lessRep[key] = expire
	lh.timer.AddTask(&less{
		key: key,
		lh:  lh,
	}, expire)
}

func (lh *LessHolder) RenewLess(instance Instance) {
	key := instance.GetKey()
	defer lh.lessLock.Unlock()
	lh.lessLock.Lock()
	if v, exist := lh.lessRep[key]; exist {
		lh.lessRep[key] = time.Duration(15)*time.Second + v
	}
}

func (lh *LessHolder) RevokeLess(instance Instance) {
	key := instance.GetKey()
	defer lh.lessLock.Unlock()
	lh.lessLock.Lock()
	delete(lh.lessRep, key)
}

func (lh *LessHolder) AddExpireWatcher(observer ExpireLessWatcher) {
	defer lh.rwLock.Unlock()
	lh.rwLock.Lock()
	lh.watchers = append(lh.watchers, observer)
}

func (lh *LessHolder) notifyExpire(keys []string) {
	defer lh.rwLock.RUnlock()
	lh.rwLock.RLock()
	for _, watcher := range lh.watchers {
		watcher(keys)
	}
}

type less struct {
	key string
	lh  *LessHolder
}

func (l *less) Run() {
	defer l.lh.rwLock.RUnlock()
	l.lh.rwLock.RLock()
	v := l.lh.lessRep[l.key]
	if int64(v) <= time.Now().Unix() {
		l.lh.expireKeys <- []string{l.key}
	} else {
		l.lh.timer.AddTask(l, v)
	}
}
