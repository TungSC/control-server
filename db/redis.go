package db

import (
    "context"
    "fmt"
    "sync"
    "time"

    "cdn/pkg/log"
)

type RedisConfig struct {
    Addrs []string `mapstructure:"addrs"`
    Pwd   string   `mapstructure:"password"`
    DB    int      `mapstructure:"db"`
}

type RedisDefaultConfig struct {
    SentinelAddrs []string `mapstructure:"sentinelAddrs"`
    Addrs         []string `mapstructure:"addrs"`
    Pwd           string   `mapstructure:"password"`
    MasterName    string   `mapstructure:"masterName"`
    DB            int      `mapstructure:"db"`
}

type RedisCenterConfig struct {
    Addrs []string `mapstructure:"addrs"`
    Pwd   string   `mapstructure:"password"`
    DB    int      `mapstructure:"db"`
}

type Redis struct {
    cluster      *redis.ClusterClient
    single       *redis.Client
    sentinel     *redis.Client
    clusterMode  bool
    sentinelMode bool
    mutex        *sync.Mutex
}

type ReceiveMessage struct {
    Action string `json:"action"`
    Server string `json:"server"`
}

func NewRedis(c RedisConfig) *Redis {
    if len(c.Addrs) == 0 {
        return nil
    }
    r := &Redis{}
    if len(c.Addrs) == 1 {
        r.single = redis.NewClient(
            &redis.Options{
                Addr: c.Addrs[0], // use default Addr
                // no password set
                Password:     c.Pwd,
                DB:           c.DB, // use default DB
                DialTimeout:  3 * time.Second,
                ReadTimeout:  5 * time.Second,
                WriteTimeout: 5 * time.Second,
            })
        if err := r.single.Ping().Err(); err != nil {
            log.Errorf(err.Error())
            return nil
        }
        r.single.Do("CONFIG", "SET", "notify-keyspace-events", "AKE")
        r.clusterMode = false
        r.mutex = new(sync.Mutex)
        fmt.Println("connection redislocal success!!!")
        return r
    }
    r.cluster = redis.NewClusterClient(
        &redis.ClusterOptions{
            Addrs:        c.Addrs,
            Password:     c.Pwd,
            DialTimeout:  3 * time.Second,
            ReadTimeout:  5 * time.Second,
            WriteTimeout: 5 * time.Second,
        })
    if err := r.cluster.Ping().Err(); err != nil {
        log.Errorf(err.Error())
    }
    r.cluster.Do("CONFIG", "SET", "notify-keyspace-events", "AKE")
    r.clusterMode = true
    return r
}

func (r *Redis) Ping() error {
    if r.clusterMode {
        err := r.cluster.Ping().Err()
        if err != nil {
            log.Errorf("[db][redisConnection][clusterMode] err: ", err)
            return err
        }
        return nil
    }
    if r.sentinelMode {
        err := r.sentinel.Ping().Err()
        if err != nil {
            log.Errorf("[db][redisConnection][sentinelMode] err: ", err)
            return err
        }
        return nil
    }
    err := r.single.Ping().Err()
    if err != nil {
        log.Errorf("[db][redisConnection][single] err: ", err)
        return err
    }
    return nil
}

func (r *Redis) Close() {
    if r.single != nil {
        r.single.Close()
    }
    if r.cluster != nil {
        r.cluster.Close()
    }
}

func (r *Redis) Set2(k string, v interface{}, t time.Duration) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Set(k, v, t).Err()
    }
    return r.single.Set(k, v, t).Err()
}

func (r *Redis) Set(k, v string, t time.Duration) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Set(k, v, t).Err()
    }
    if r.sentinelMode {
        return r.sentinel.Set(k, v, t).Err()
    }
    return r.single.Set(k, v, t).Err()
}

func (r *Redis) Get(k string) string {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Get(k).Val()
    }
    if r.sentinelMode {
        return r.sentinel.Get(k).Val()
    }
    return r.single.Get(k).Val()
}

func (r *Redis) HSet(k, field string, value interface{}) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.HSet(k, field, value).Err()
    }
    if r.sentinelMode {
        return r.sentinel.HSet(k, field, value).Err()
    }
    return r.single.HSet(k, field, value).Err()
}

func (r *Redis) HMSet(k string, value interface{}) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.HMSet(k, value).Err()
    }
    if r.sentinelMode {
        return r.sentinel.HMSet(k, value).Err()
    }
    return r.single.HMSet(k, value).Err()
}

func (r *Redis) HGet(k, field string) string {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.HGet(k, field).Val()
    }
    if r.sentinelMode {
        return r.sentinel.HGet(k, field).Val()
    }
    return r.single.HGet(k, field).Val()
}

func (r *Redis) HGetAll(k string) map[string]string {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.HGetAll(k).Val()
    }
    if r.sentinelMode {
        return r.sentinel.HGetAll(k).Val()
    }
    return r.single.HGetAll(k).Val()
}

func (r *Redis) HDel(k, field string) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.HDel(k, field).Err()
    }
    if r.sentinelMode {
        return r.sentinel.HDel(k, field).Err()
    }
    return r.single.HDel(k, field).Err()
}

func (r *Redis) Expire(k string, t time.Duration) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Expire(k, t).Err()
    }
    if r.sentinelMode {
        return r.sentinel.Expire(k, t).Err()
    }
    return r.single.Expire(k, t).Err()
}

func (r *Redis) HSetTTL(k, field string, value interface{}, t time.Duration) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        if err := r.cluster.HSet(k, field, value).Err(); err != nil {
            return err
        }
        return r.cluster.Expire(k, t).Err()
    }
    if r.sentinelMode {
        if err := r.sentinel.HSet(k, field, value).Err(); err != nil {
            return err
        }
        return r.sentinel.Expire(k, t).Err()
    }
    if err := r.single.HSet(k, field, value).Err(); err != nil {
        return err
    }
    return r.single.Expire(k, t).Err()
}

func (r *Redis) Keys(k string) []string {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Keys(k).Val()
    }
    if r.sentinelMode {
        return r.sentinel.Keys(k).Val()
    }
    return r.single.Keys(k).Val()
}

func (r *Redis) Del(k string) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Del(k).Err()
    }
    if r.sentinelMode {
        return r.sentinel.Del(k).Err()
    }
    return r.single.Del(k).Err()
}

func (r *Redis) Publish(channel string, message interface{}) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Publish(channel, message).Err()
    }
    if r.sentinelMode {
        return r.sentinel.Publish(channel, message).Err()
    }
    return r.single.Publish(channel, message).Err()
}

func (r *Redis) Subscribe(k string) *redis.PubSub {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    if r.clusterMode {
        return r.cluster.Subscribe(k)
    }
    if r.sentinelMode {
        return r.sentinel.Subscribe(k)
    }
    return r.single.Subscribe(k)
}

// Watch http://redisdoc.com/topic/notification.html
func (r *Redis) Watch(ctx context.Context, key string) <-chan interface{} {
    var pubsub *redis.PubSub
    pubsub = r.single.PSubscribe(key)
    if r.clusterMode {
        pubsub = r.cluster.PSubscribe(key)
    }
    if r.sentinelMode {
        pubsub = r.sentinel.PSubscribe(key)
    }

    res := make(chan interface{})
    go func() {
        for {
            select {
            case msg := <-pubsub.Channel():
                op := msg.Payload
                log.Infof("key => %s, op => %s", key, op)
                res <- op
            case <-ctx.Done():
                pubsub.Close()
                close(res)
                return
            }
        }
    }()

    return res
}
