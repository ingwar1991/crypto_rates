package auth_connections

import (
	"context"
	"fmt"
	"sync"
	"time"

	mongo_helper "crypto_rates_auth/internal/mongo"
)


type ActiveConnections struct {
	ctx context.Context

	mu sync.RWMutex 

	active map[string]*mongo_helper.ActiveSecret
}

func NewActiveConnections(ctx context.Context) *ActiveConnections {
	cs := &ActiveConnections{
		ctx: ctx,
		active: make(map[string]*mongo_helper.ActiveSecret),
	}

	go func() {
		ticker := time.NewTicker(1*time.Minute) 
		defer ticker.Stop()

		for {
			select {
			case <- ticker.C:
				cs.Cleanup()
			case <- cs.Ctx().Done():
				return
			}
		}
	}()

	return cs
}

func (cs *ActiveConnections) Ctx() context.Context {
	return cs.ctx
}

func (cs *ActiveConnections) Get(secret string) (*mongo_helper.ActiveSecret, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	conn, ok := cs.active[secret]

	return conn, ok
}

func (cs *ActiveConnections) Add(conn *mongo_helper.ActiveSecret) error {
	if !conn.IsValid(time.Now()) { 
		return fmt.Errorf("[connections.add] can not add expired secret: %v", conn)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.active[conn.Secret] = conn

	return nil
}

func (cs *ActiveConnections) Remove(secret string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.active, secret)
}

func (cs *ActiveConnections) Check(secret string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	conn, ok := cs.active[secret]
	if !ok {
		return false
	}
	
	return conn.IsValid(time.Now())
}

func (cs *ActiveConnections) collectForCleanup() (listToRemove []string) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	tNow := time.Now()
	for key, val := range cs.active {
		if !val.IsValid(tNow) {
			listToRemove = append(listToRemove, key)
		}
	}

	return listToRemove
}

func (cs *ActiveConnections) Cleanup() {
	listToRemove := cs.collectForCleanup()
	if len(listToRemove) < 1 {
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, key := range listToRemove {
		delete(cs.active, key)
	}
}
