package client

import (
	"goto/store"
	"log"
	"net/rpc"
)

type ProxyStore struct {
	urls   *store.URLStore
	client *rpc.Client
}

func NewProxyStore(addr string) *ProxyStore {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		log.Println("Error constructing ProxyStore:", err)
	}

	proxy := &ProxyStore{store.NewURLStore(""), client}
	go func() {
		data := make(map[string]string)
		u := 0
		err_ := proxy.Sync(&u, &data)
		if err_ == nil {
			for k, v := range data {
				proxy.urls.Set(&k, &v)
			}
			log.Println("rpcClient data synced，count", len(data))
		} else {
			log.Println("rpcClient data sync error", err_)
		}
	}()
	return proxy
}

func (s *ProxyStore) Get(key, url *string) error {
	log.Printf("rpcClient get:{%s}\n", *key)
	if err := s.urls.Get(key, url); err == nil {
		log.Printf("rpcClient get:{%s} load from client cache\n", *key)
		return nil
	}
	if err := s.client.Call("Store.Get", key, url); err != nil {
		return err
	}
	return s.urls.Set(key, url)
}

func (s *ProxyStore) Put(url, key *string) error {
	if err := s.client.Call("Store.Put", url, key); err != nil {
		return err
	}
	log.Println("rpcClient put:{%s},{%s}", *url, *key)
	return s.urls.Set(key, url)
}
func (s *ProxyStore) Sync(u *int, urls *map[string]string) error {
	if err := s.client.Call("Store.Sync", u, urls); err != nil {
		return err
	}
	return nil
}
