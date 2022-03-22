// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/3/29

package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GroupNewElementFunc 给 Group 创建新的 pool
type GroupNewElementFunc func(key interface{}) NewElementFunc

// NewSimplePoolGroup 创建新的 Group
func NewSimplePoolGroup(opt *Option, gn GroupNewElementFunc) SimplePoolGroup {
	if opt == nil {
		opt = &Option{}
	}
	ctx, cancel := context.WithCancel(context.Background())

	sgOpt := opt.Clone()
	sgOpt.MaxLifeTime = 0 // 避免由于生命周期被强制关闭

	// 设置一个更合适的 idle 时间，避免子 pool 被清理掉
	minIdle := 3 * time.Minute
	if sgOpt.MaxIdleTime < minIdle {
		sgOpt.MaxIdleTime = minIdle
	}

	g := &simpleGroup{
		rawOption: *opt.Clone(),
		sgOption:  *sgOpt,
		done:      cancel,
		genNewEle: gn,
	}
	go g.poolCleaner(ctx, opt.shortestIdleTime())
	return g
}

// SimplePoolGroup 通用的、 按照 key 分组的 Pool
type SimplePoolGroup interface {
	Get(ctx context.Context, key interface{}) (Element, error)
	GroupStats() GroupStats
	Close() error
	Option() Option
	Range(func(el Element) error) error
}

var _ SimplePoolGroup = (*simpleGroup)(nil)

// simpleGroup group pool
type simpleGroup struct {
	rawOption Option // 用于传递给子 Pool 的 option

	sgOption Option // 用来判断子 pool 状态的 option

	genNewEle GroupNewElementFunc
	pools     map[interface{}]*groupPoolItem
	mu        sync.Mutex
	done      context.CancelFunc
	closed    bool
}

func (g *simpleGroup) Range(fn func(el Element) error) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, pool := range g.pools {
		if err := pool.Range(fn); err != nil {
			return err
		}
	}
	return nil
}

func (g *simpleGroup) Option() Option {
	return g.rawOption
}

// Get ...
func (g *simpleGroup) Get(ctx context.Context, key interface{}) (Element, error) {
	return g.getPool(key).Get(ctx)
}

func (g *simpleGroup) getPool(key interface{}) *groupPoolItem {
	poolID := getPoolID(key)
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.pools == nil {
		g.pools = make(map[interface{}]*groupPoolItem)
	}
	p, has := g.pools[poolID]
	if !has {
		fn := g.genNewEle(key)
		pool := NewSimplePool(&g.rawOption, fn)
		p = newGroupPoolItem(pool)
		g.pools[poolID] = p
	}
	p.PEMarkUsing()
	return p
}

// GroupStats Group 的状态信息
func (g *simpleGroup) GroupStats() GroupStats {
	g.mu.Lock()
	defer g.mu.Unlock()

	gs := GroupStats{
		Groups: make([]*GroupStatDetail, 0, len(g.pools)),
		All: Stats{
			Open: !g.closed,
		},
	}

	if g.pools == nil {
		return gs
	}

	for key, p := range g.pools {
		ls := p.Stats()
		detail := &GroupStatDetail{
			Group: key,
			Stats: ls,
		}
		gs.Groups = append(gs.Groups, detail)

		gs.All.Idle += ls.Idle
		gs.All.NumOpen += ls.NumOpen
		gs.All.InUse += ls.InUse
		gs.All.WaitCount += ls.WaitCount
		gs.All.WaitDuration += ls.WaitDuration
		gs.All.MaxIdleClosed += ls.MaxIdleClosed
		gs.All.MaxIdleTimeClosed += ls.MaxIdleTimeClosed
		gs.All.MaxLifeTimeClosed += ls.MaxLifeTimeClosed
	}
	return gs
}

// Close close pools
func (g *simpleGroup) Close() error {
	g.done()

	var err error
	g.mu.Lock()
	g.closed = true

	if g.pools != nil {
		for _, p := range g.pools {
			if e := p.Close(); e != nil {
				err = e
			}
		}
		g.pools = make(map[interface{}]*groupPoolItem)
	}
	g.mu.Unlock()

	return err
}

// poolCleaner 对连接池分组进行检查，删除无效的，不再使用的连接池
func (g *simpleGroup) poolCleaner(ctx context.Context, d time.Duration) {
	const minInterval = time.Minute

	if d < minInterval {
		d = minInterval
	}
	t := time.NewTimer(d)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		g.doCheckExpire()
		t.Reset(d)
	}
}

func (g *simpleGroup) doCheckExpire() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.pools == nil {
		return
	}
	var expires []interface{}
	for k, p := range g.pools {
		if err := p.Active(g.sgOption); err != nil {
			expires = append(expires, k)
			p.Close()
		}
	}
	if len(expires) == 0 {
		return
	}

	for _, k := range expires {
		delete(g.pools, k)
	}
}

type groupPoolItem struct {
	*MetaInfo
	SimplePool
}

func newGroupPoolItem(p SimplePool) *groupPoolItem {
	return &groupPoolItem{
		MetaInfo:   NewMetaInfo(),
		SimplePool: p,
	}
}

func getPoolID(key interface{}) interface{} {
	if v, ok := key.(interface{ PoolID() interface{} }); ok {
		return v.PoolID()
	}

	if v, ok := key.(fmt.Stringer); ok {
		return v.String()
	}

	return key
}
