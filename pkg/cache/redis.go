// Package cache bọc go-redis thành một API nhỏ, đủ dùng cho các service.
//
// Nguyên tắc áp dụng trong repo này: Redis là LỚP TĂNG TỐC, không phải nguồn
// sự thật. Mọi hàm Get đều "fail-open" — Redis chết thì coi như cache miss và
// đi tiếp xuống Postgres, chứ không được làm hỏng request.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss là lỗi "không có trong cache". Đây là lỗi BÌNH THƯỜNG.
var ErrCacheMiss = errors.New("cache: key not found")

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
	// Prefix để nhiều service dùng chung 1 Redis mà không giẫm chân nhau.
	Prefix string
}

func (c Config) Addr() string { return fmt.Sprintf("%s:%s", c.Host, c.Port) }

// Client là handle cache của một service.
type Client struct {
	rdb    *redis.Client
	prefix string
}

// New mở kết nối và ping thử. Ping lỗi thì trả về error để DI quyết định:
// hoặc dừng service, hoặc chạy tiếp ở chế độ không cache.
func New(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     20,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: cannot reach redis at %s: %w", cfg.Addr(), err)
	}

	return &Client{rdb: rdb, prefix: cfg.Prefix}, nil
}

// Raw trả về client gốc cho các lệnh đặc thù (GEO, PUBSUB...).
func (c *Client) Raw() *redis.Client { return c.rdb }

// Key ghép prefix với các thành phần. Nil-safe: repo dựng key trước rồi mới
// kiểm tra cache có tồn tại hay không, nên hàm này phải chịu được receiver nil.
func (c *Client) Key(parts ...string) string {
	if c == nil {
		return strings.Join(parts, ":")
	}
	key := c.prefix
	for _, p := range parts {
		key += ":" + p
	}
	return key
}

func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// SetJSON ghi giá trị dạng JSON kèm TTL.
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil {
		return nil
	}
	blob, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal %s: %w", key, err)
	}
	return c.rdb.Set(ctx, key, blob, ttl).Err()
}

// GetJSON đọc và unmarshal. Trả ErrCacheMiss khi không có key.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	if c == nil {
		return ErrCacheMiss
	}
	blob, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		// Redis lỗi => coi như miss, để caller đi xuống DB.
		log.Printf("[cache] GET %s failed: %v (treating as miss)", key, err)
		return ErrCacheMiss
	}
	if err := json.Unmarshal(blob, dest); err != nil {
		// Cache bẩn (đổi struct chẳng hạn) => xoá đi cho lần sau ghi lại.
		_ = c.Delete(ctx, key)
		return ErrCacheMiss
	}
	return nil
}

// Delete xoá 1 hoặc nhiều key. Dùng khi ghi DB để invalidate.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if c == nil || len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// DeleteByPattern xoá theo pattern bằng SCAN (KHÔNG dùng KEYS vì KEYS block
// toàn bộ Redis single-thread — đủ để kéo sập production khi keyspace lớn).
func (c *Client) DeleteByPattern(ctx context.Context, pattern string) error {
	if c == nil {
		return nil
	}
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

// Incr tăng bộ đếm và đặt TTL nếu key vừa được tạo.
func (c *Client) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if c == nil {
		return 0, ErrCacheMiss
	}
	pipe := c.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// Decr giảm bộ đếm nhưng chặn không cho xuống dưới 0.
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	if c == nil {
		return 0, ErrCacheMiss
	}
	v, err := c.rdb.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if v < 0 {
		c.rdb.Set(ctx, key, 0, 0)
		return 0, nil
	}
	return v, nil
}

// ---------------------------------------------------------------------------
// GEO — trái tim của "tìm xe gần đây"
// ---------------------------------------------------------------------------
// Redis GEO thực chất là sorted set với score = geohash 52-bit. Nhờ đó
// GEOSEARCH quét bán kính trong O(log N + M) ngay trên RAM, thay vì bắt Postgres
// chạy ST_DWithin cho từng lần tài xế ping GPS (vài giây một lần, hàng nghìn xe).

type GeoMember struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type GeoHit struct {
	Name       string
	DistanceKm float64
	Latitude   float64
	Longitude  float64
}

// GeoAdd ghi/cập nhật toạ độ hiện tại của một phương tiện.
func (c *Client) GeoAdd(ctx context.Context, key string, members ...GeoMember) error {
	if c == nil || len(members) == 0 {
		return nil
	}
	locations := make([]*redis.GeoLocation, 0, len(members))
	for _, m := range members {
		locations = append(locations, &redis.GeoLocation{
			Name:      m.Name,
			Latitude:  m.Latitude,
			Longitude: m.Longitude,
		})
	}
	return c.rdb.GeoAdd(ctx, key, locations...).Err()
}

// GeoSearch tìm các member nằm trong bán kính radiusKm, sắp xếp gần -> xa.
func (c *Client) GeoSearch(ctx context.Context, key string, lat, lng, radiusKm float64, limit int) ([]GeoHit, error) {
	if c == nil {
		return nil, ErrCacheMiss
	}
	res, err := c.rdb.GeoSearchLocation(ctx, key, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      limit,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil {
		return nil, err
	}

	hits := make([]GeoHit, 0, len(res))
	for _, r := range res {
		hits = append(hits, GeoHit{
			Name:       r.Name,
			DistanceKm: r.Dist,
			Latitude:   r.Latitude,
			Longitude:  r.Longitude,
		})
	}
	return hits, nil
}

// GeoRemove gỡ xe khỏi index khi tài xế offline / xe vào bảo dưỡng.
func (c *Client) GeoRemove(ctx context.Context, key string, members ...string) error {
	if c == nil || len(members) == 0 {
		return nil
	}
	args := make([]any, 0, len(members))
	for _, m := range members {
		args = append(args, m)
	}
	return c.rdb.ZRem(ctx, key, args...).Err()
}
