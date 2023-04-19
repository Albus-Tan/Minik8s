package apiserver

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

var (
	etcdEndpoint   = "localhost:2379"
	requestTimeout = time.Second
	etcdConfig     clientv3.Config
	etcdClient     *clientv3.Client
)

func initEtcd() {
	etcdConfig = clientv3.Config{
		Endpoints:            []string{etcdEndpoint},
		DialTimeout:          30 * time.Second,
		DialKeepAliveTimeout: 30 * time.Second,
	}

	var err error
	etcdClient, err = clientv3.New(etcdConfig)
	if err != nil {
		log.Printf("[etcd] connect to etcd failed, err:%v\n", err)
	} else {
		log.Printf("[etcd] connect to etcd success\n")
	}

}

func closeEtcd() {
	err := etcdClient.Close()
	if err != nil {
		log.Printf("[etcd] close etcd client failed, err:%v\n", err)
	} else {
		log.Printf("[etcd] etcd client closed\n")
	}
}

func etcdPut(key, value string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = etcdClient.Put(ctx, key, value)
	cancel()
	if err != nil {
		log.Printf("[etcd] Put failed, err:%v\n", err)
		switch err {
		case context.Canceled:
			fmt.Printf("[etcd] ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			fmt.Printf("[etcd] ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			fmt.Printf("[etcd] client-side error: %v\n", err)
		default:
			fmt.Printf("[etcd] bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
	}
	return err
}

func etcdGet(key string) (value string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	if err != nil {
		log.Printf("[etcd] Get failed, err:%v\n", err)
		return "", err
	}

	if len(resp.Kvs) > 0 {
		return string(resp.Kvs[0].Value), err
	} else {
		return "", err
	}
}

func etcdHas(key string) (value bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	if err != nil {
		log.Printf("[etcd] Has failed, err:%v\n", err)
		return false, err
	}

	if resp.Count == 0 {
		return false, err
	} else {
		return true, err
	}
}

func etcdGetAllWithPrefix(keyPrefix string) (values []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, keyPrefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		log.Printf("[etcd] GetAllWithPrefix failed, err:%v\n", err)
		return nil, err
	}
	for _, ev := range resp.Kvs {
		values = append(values, string(ev.Value))
	}
	return values, err
}

func etcdDelete(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = etcdClient.Delete(ctx, key)
	cancel()

	if err != nil {
		log.Printf("[etcd] Delete failed, err:%v\n", err)
	}
	return err
}

func etcdDeleteAllWithPrefix(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = etcdClient.Delete(ctx, key, clientv3.WithPrefix())
	cancel()

	if err != nil {
		log.Printf("[etcd] DeleteAll failed, err:%v\n", err)
	}
	return err
}

func etcdClear() (err error) {
	return etcdDeleteAllWithPrefix("")
}
