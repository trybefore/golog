package loadbalance

import (
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

func init() {
	balancer.Register(base.NewBalancerBuilder(Name, &Picker{}, base.Config{}))
}

var _ base.PickerBuilder = (*Picker)(nil)
var _ balancer.Picker = (*Picker)(nil)

type Picker struct {
	mtx    sync.RWMutex
	leader balancer.SubConn

	followers []balancer.SubConn
	current   uint64
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	var result balancer.PickResult
	if strings.Contains(info.FullMethodName, "Produce") || len(p.followers) == 0 {
		result.SubConn = p.leader
	} else if strings.Contains(info.FullMethodName, "Consume") {
		result.SubConn = p.nextFollower()
	}
	if result.SubConn == nil {
		return result, balancer.ErrNoSubConnAvailable
	}

	return result, nil
}

func (p *Picker) nextFollower() balancer.SubConn {
	cur := atomic.AddUint64(&p.current, uint64(1))
	len := uint64(len(p.followers))
	idx := int(cur % len)

	return p.followers[idx]
}

func (p *Picker) Build(info base.PickerBuildInfo) balancer.Picker {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	var followers []balancer.SubConn
	for sc, scInfo := range info.ReadySCs {
		isLeader := scInfo.Address.Attributes.Value("is_leader").(bool)

		if isLeader {
			p.leader = sc
			continue
		}
		followers = append(followers, sc)
	}
	p.followers = followers
	return p
}
