package repository

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

// все ключи логов начинаются с этого
const logPrefix = "log:"

// логи изменений хранятся в редисе, у каждой записи свой срок жизни
type RedisLogger struct {
	client *redis.Client
	ttl    time.Duration
}

func ConnectRedis(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := newCtx()
	defer cancel()
	err := client.Ping(ctx).Err()
	if err != nil {
		// редис не ответил, клиент больше не нужен
		client.Close()
		return nil, err
	}
	return client, nil
}

func NewRedisLogger(client *redis.Client, ttl time.Duration) *RedisLogger {
	return &RedisLogger{client: client, ttl: ttl}
}

type logEntry struct {
	Time   string `json:"time"`
	Action string `json:"action"`
	Data   any    `json:"data"`
}

func (l *RedisLogger) Log(action string, data any) {
	now := time.Now()
	bytes, err := json.Marshal(logEntry{Time: now.Format(time.RFC3339), Action: action, Data: data})
	if err != nil {
		fmt.Println("не удалось сделать json для лога:", err)
		return
	}

	// время в ключе чтобы потом можно было отсортировать
	key := fmt.Sprintf("%s%d", logPrefix, now.UnixNano())

	ctx, cancel := newCtx()
	defer cancel()
	err = l.client.Set(ctx, key, bytes, l.ttl).Err()
	if err != nil {
		fmt.Println("не удалось записать лог в redis:", err)
	}
}

// все логи которые ещё не протухли, старые сначала
func (l *RedisLogger) All() []string {
	ctx, cancel := newCtx()
	defer cancel()

	// SCAN вместо KEYS, чтобы не блокировать редис на большом количестве ключей
	keys := []string{}
	iter := l.client.Scan(ctx, 0, logPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		fmt.Println("не удалось получить логи из redis:", err)
		return []string{}
	}
	list := []string{}
	if len(keys) == 0 {
		return list
	}
	sort.Strings(keys)

	// забираем все значения одним запросом
	vals, err := l.client.MGet(ctx, keys...).Result()
	if err != nil {
		fmt.Println("не удалось получить логи из redis:", err)
		return list
	}
	for _, val := range vals {
		s, ok := val.(string)
		if !ok {
			// мог протухнуть пока читали
			continue
		}
		list = append(list, s)
	}
	return list
}
