package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"minik8s/config"
	"minik8s/pkg/logger"
	"strconv"
	"sync"
	"time"
)

type Event clientv3.Event

const (
	EventTypeDelete = clientv3.EventTypeDelete
	EventTypePut    = clientv3.EventTypePut
)

var (
	etcdEndpoint   = config.EtcdHost + config.EtcdPort
	requestTimeout = time.Second
	etcdConfig     clientv3.Config
	etcdClient     *clientv3.Client
)

const EmptyGetResult string = ""

// Rvm store and manage global ResourceVersion
var Rvm ResourceVersionManager

type ResourceVersionManager struct {
	version int64
	mutex   sync.RWMutex
}

func (r *ResourceVersionManager) GetNextResourceVersion() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	logger.ApiServerLogger.Printf("[ResourceVersionManager] GetNextResourceVersion %v\n", r.version)
	return strconv.FormatInt(r.version+1, 10)
}

func (r *ResourceVersionManager) GetResourceVersion() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	logger.ApiServerLogger.Printf("[ResourceVersionManager] GetResourceVersion %v\n", r.version)
	return strconv.FormatInt(r.version, 10)
}

func (r *ResourceVersionManager) setResourceVersion(v string) {
	r.mutex.Lock()
	r.version, _ = strconv.ParseInt(v, 10, 64)
	logger.ApiServerLogger.Printf("[ResourceVersionManager] SetResourceVersion %v\n", r.version)
	r.mutex.Unlock()
}

func (r *ResourceVersionManager) init(v int64) {
	r.mutex.Lock()
	r.version = v
	logger.ApiServerLogger.Printf("[ResourceVersionManager] init version %v\n", r.version)
	r.mutex.Unlock()
}

func Init() {
	etcdConfig = clientv3.Config{
		Endpoints:            []string{etcdEndpoint},
		DialTimeout:          30 * time.Second,
		DialKeepAliveTimeout: 30 * time.Second,
	}

	var err error
	etcdClient, err = clientv3.New(etcdConfig)
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] connect to etcd failed, err:%v\n", err)
	} else {
		logger.ApiServerLogger.Printf("[etcd] connect to etcd success\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	status, err := etcdClient.Status(ctx, etcdEndpoint)
	cancel()
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] get etcdClient status failed, err:%v\n", err)
	}

	Rvm.init(status.Header.Revision)
}

func Close() {
	err := etcdClient.Close()
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] close etcd client failed, err:%v\n", err)
	} else {
		logger.ApiServerLogger.Printf("[etcd] etcd client closed\n")
	}
}

func Put(key, value string) (err error, newVersion string) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Put(ctx, key, value)
	cancel()
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] Put failed, err:%v\n", err)
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
	newVersion = strconv.FormatInt(resp.Header.Revision, 10)
	Rvm.setResourceVersion(newVersion)
	fmt.Printf("[etcd] Put: newVersion %v, resp %v\n", newVersion, resp)

	return err, newVersion
}

func CheckVersionPut(key, value, oldVersion string) (err error, newVersion string, success bool) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Put(ctx, key, value, clientv3.WithPrevKV())
	cancel()
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] Put failed, err:%v\n", err)
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

	// check version
	newVersion = strconv.FormatInt(resp.Header.Revision, 10)
	Rvm.setResourceVersion(newVersion)
	fmt.Printf("[etcd] CheckVersionPut: newVersion %v, oldVersion %v, resp.PrevKv.ModRevision %v, resp %v\n", newVersion, oldVersion, resp.PrevKv.ModRevision, resp)

	if oldVersion != strconv.FormatInt(resp.PrevKv.ModRevision, 10) {
		fmt.Printf("[etcd] CheckVersionPut FAILED, oldVersion %v and resp.PrevKv.ModRevision %v mismatch\n", oldVersion, resp.PrevKv.ModRevision)
		return err, newVersion, false
	}
	fmt.Printf("[etcd] CheckVersionPut SUCCESS\n")
	return err, newVersion, true
}

func Get(key string) (value string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	fmt.Printf("[etcd] Get: resp %v\n", resp)

	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] Get failed, err:%v\n", err)
		return EmptyGetResult, err
	}

	if len(resp.Kvs) > 0 {
		return string(resp.Kvs[0].Value), err
	} else {
		return EmptyGetResult, err
	}
}

func GetWithVersion(key string) (value string, version string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	fmt.Printf("[etcd] GetWithVersion: resp %v\n", resp)

	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] Get failed, err:%v\n", err)
		return EmptyGetResult, version, err
	}

	if len(resp.Kvs) > 0 {
		version = strconv.FormatInt(resp.Kvs[0].ModRevision, 10)
		return string(resp.Kvs[0].Value), version, err
	} else {
		return EmptyGetResult, version, err
	}
}

func Has(key string) (value bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] Has failed, err:%v\n", err)
		return false, err
	}

	if resp.Count == 0 {
		return false, err
	} else {
		return true, err
	}
}

func HasWithVersion(key string) (value bool, version string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()

	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] HasWithVersion failed, err:%v\n", err)
		return false, version, err
	}

	if resp.Count == 0 {
		return false, version, err
	} else {
		version = strconv.FormatInt(resp.Kvs[0].ModRevision, 10)
		return true, version, err
	}
}

func GetAllWithPrefix(keyPrefix string) (values []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Get(ctx, keyPrefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] GetAllWithPrefix failed, err:%v\n", err)
		return nil, err
	}
	for _, ev := range resp.Kvs {
		values = append(values, string(ev.Value))
	}
	return values, err
}

func Delete(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Delete(ctx, key)
	cancel()
	Rvm.setResourceVersion(strconv.FormatInt(resp.Header.Revision, 10))

	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] Delete failed, err:%v\n", err)
	}
	return err
}

func DeleteAllWithPrefix(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := etcdClient.Delete(ctx, key, clientv3.WithPrefix())
	cancel()
	Rvm.setResourceVersion(strconv.FormatInt(resp.Header.Revision, 10))

	if err != nil {
		logger.ApiServerLogger.Printf("[etcd] DeleteAll failed, err:%v\n", err)
	}
	return err
}

func Clear() (err error) {
	return DeleteAllWithPrefix("")
}

func Watch(key string) (context.CancelFunc, chan *Event) {
	ctx, cancel := context.WithCancel(context.Background())
	rch := etcdClient.Watch(ctx, key)
	ch := make(chan *Event)
	go doWatch(rch, ch)
	return cancel, ch
}

func WatchAllWithPrefix(key string) (context.CancelFunc, chan *Event) {
	ctx, cancel := context.WithCancel(context.Background())
	rch := etcdClient.Watch(ctx, key, clientv3.WithPrefix())
	ch := make(chan *Event)
	go doWatch(rch, ch)
	return cancel, ch
}

func doWatch(rch clientv3.WatchChan, ch chan *Event) {
	// continue to read rch until it's closed
	for wresp := range rch {
		for _, ev := range wresp.Events {
			logger.ApiServerLogger.Printf("[etcd] watch notified %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			ch <- (*Event)(ev)
		}
	}
}
