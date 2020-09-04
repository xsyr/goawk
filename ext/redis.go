package ext

import (
    "github.com/xsyr/goawk/interp"
    "log"
    "strings"
    "strconv"
    "github.com/go-redis/redis"
)


type redisModule struct {
    opt redis.Options
    funcs []*FuncDef
}

func newRedisModule() Module {
    var m = &redisModule{
        opt : redis.Options{
            Addr:     "localhost:6379",
            Password: "", // no password set
            DB:       0,  // use default DB
        },
    }
    m.define()
    return m
}

func (m *redisModule) Desc() string {
    return "redis"
}

func (m *redisModule) Funcs() []*FuncDef {
    return m.funcs
}

func (m *redisModule)SetRuntime(rt *interp.Runtime) {

}

func (m *redisModule) define() {
    m.funcs = []*FuncDef{
        {
            Desc: "__redis_addr(host). default localhost:6379",
            Func: func(host string) {
                m.opt.Addr = host
            },
        },
        {
            Desc: "__redis_password(port). default empty",
            Func: func(pwd string) {
                m.opt.Password = pwd
            },
        },
        {
            Desc: "__redis_db(db). default 0",
            Func: func(db int) {
                m.opt.DB = db
            },
        },
        {
            Desc: "__redis_cmd(args...) interface{}",
            Func: func(args ...string) interface{} {
                client := redis.NewClient(&m.opt)
                defer client.Close()

                as := make([]interface{}, len(args))
                for i, arg := range args {
                    as[i] = arg
                }
                val, err := client.Do(as...).Result()
                if err == redis.Nil {
                    return ""
                }
                if err != nil {
                    log.Printf("redis.Do() failed: %v", err)
                    return ""
                }
                switch vl1 := val.(type) {
                case string:
                    return vl1
                case int64:
                    return strconv.FormatInt(vl1, 10)
                case []string:
                    return vl1
                case []interface{}:
                    {
                        if len(vl1) == 0 {
                            return ""
                        }

                        vs := make([]string, len(vl1))
                        switch vl1[0].(type) {
                        case string:
                            for i, v := range vl1 {
                                vs[i] = v.(string)
                            }
                            return vs

                        case int64:
                            for i, v := range vl1 {
                                vs[i] = strconv.FormatInt(v.(int64), 10)
                            }
                            return vs
                        case nil:
                            return vs
                        default:
                            log.Printf("can't handle type=[%T], val=[%v]\n", val, val)
                        }
                    }

                default:
                    log.Printf("can't handle type=[%T], val=[%v]\n", val, val)
                }
                return ""
            },
        },
    }

    for _, fn := range m.funcs {
        fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
    }
}