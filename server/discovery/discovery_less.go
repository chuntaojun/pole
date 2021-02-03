package discovery

import (
	"context"
	"sync"
	"time"

	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/server/sys"
	"github.com/pole-group/pole/utils"
)

type ExpireLessWatcher func(keys []string)

type Lessor struct {
	ctx        *common.ContextPole
	lessLock   sync.RWMutex
	lessRep    map[string]time.Duration
	rwLock     sync.RWMutex
	expireKeys chan []string
	watchers   []ExpireLessWatcher
	timer      *polerpc.HashTimeWheel
}

func NewLessor(ctx *common.ContextPole) *Lessor {
	lh := &Lessor{
		ctx:      ctx,
		lessLock: sync.RWMutex{},
		lessRep:  make(map[string]time.Duration),
		rwLock:   sync.RWMutex{},
		watchers: make([]ExpireLessWatcher, 0, 0),
		timer: polerpc.NewTimeWheel(func(opts *polerpc.Options) {
			opts.SlotNum = 15
			opts.Interval = time.Duration(1) * time.Second
			opts.MaxDealTask = 65535
		}),
	}

	lh.start()
	return lh
}

func (lh *Lessor) start() {
	polerpc.Go(lh.ctx.NewSubCtx(), func(ctx context.Context) {
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

func (lh *Lessor) GrantLess(instance Instance) {
	key := instance.GetKey()
	defer lh.lessLock.Unlock()
	lh.lessLock.Lock()
	expire := time.Duration(15+utils.GetCurrentTimeMs()) * time.Second
	lh.lessRep[key] = expire
	lh.timer.DelayExec(&less{
		key: key,
		lh:  lh,
	}, expire)
}

func (lh *Lessor) RenewLess(instance Instance) {
	key := instance.GetKey()
	defer lh.lessLock.Unlock()
	lh.lessLock.Lock()
	if v, exist := lh.lessRep[key]; exist {
		lh.lessRep[key] = time.Duration(15)*time.Second + v
	}
}

func (lh *Lessor) RevokeLess(instance Instance) {
	key := instance.GetKey()
	defer lh.lessLock.Unlock()
	lh.lessLock.Lock()
	delete(lh.lessRep, key)
}

func (lh *Lessor) AddExpireWatcher(observer ExpireLessWatcher) {
	defer lh.rwLock.Unlock()
	lh.rwLock.Lock()
	lh.watchers = append(lh.watchers, observer)
}

func (lh *Lessor) notifyExpire(keys []string) {
	defer lh.rwLock.RUnlock()
	lh.rwLock.RLock()
	for _, watcher := range lh.watchers {
		watcher(keys)
	}
}

type less struct {
	key string
	lh  *Lessor
}

func (l *less) Run() {
	defer l.lh.rwLock.RUnlock()
	l.lh.rwLock.RLock()
	v, ok := l.lh.lessRep[l.key]
	if !ok {
		sys.DiscoveryLessLogger.Warn("instance : %s less info not found, "+
			"so canceled this check task", l.key)
		return
	}
	//为了避免由于调用时间陷入系统调用所带来的额外开销，这里使用一个定时任务去定期的查询时间信息
	if int64(v) <= utils.GetCurrentTimeMs() {
		l.lh.expireKeys <- []string{l.key}
	} else {
		l.lh.timer.DelayExec(l, v)
	}
}
