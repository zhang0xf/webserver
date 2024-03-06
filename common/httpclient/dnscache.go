package httpclient

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type DnsCacheItem struct {
	IP        string
	CacheTime int64
}

type DnsCache struct {
	sync.RWMutex
	caches map[string]DnsCacheItem
}

// the cache will not remove without a trigger of http get
func (dnsCache *DnsCache) Get(addr string) string {
	if DnsCacheDuration <= 0 {
		return addr
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	dnsCache.RLock()
	item, ok := dnsCache.caches[host]
	dnsCache.RUnlock()

	if !ok || time.Now().Unix()-item.CacheTime > int64(DnsCacheDuration/time.Second) {
		go func() {
			netAddr, err := net.ResolveTCPAddr("tcp", addr)
			if err == nil {
				dnsCache.Lock()
				dnsCache.caches[host] = DnsCacheItem{IP: netAddr.IP.String(), CacheTime: time.Now().Unix()}
				dnsCache.Unlock()
			}
		}()
	}
	if ok {
		return fmt.Sprintf("%s:%s", item.IP, port)
	} else {
		return addr
	}
}
