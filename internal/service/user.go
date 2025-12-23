package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"golang.org/x/sync/singleflight"

	"github.com/test-tt/internal/dao"
	"github.com/test-tt/internal/model"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/logger"
)

const (
	userCacheKey      = "user:%d"
	userPageCacheKey  = "users:page:%d:%d" // page:pageSize
	userCountCacheKey = "users:count"
	userListCacheKey  = "users:all"
	cacheTTL          = 5 * time.Minute
	localCacheTTL     = 30 * time.Second // 本地缓存 TTL（短于 Redis，避免数据不一致）
	countCacheTTL     = 1 * time.Minute  // count 缓存时间短一些
)

// singleflight 防止缓存击穿
var sf singleflight.Group

type UserService struct {
	userDAO *dao.UserDAO
}

func NewUserService() *UserService {
	return &UserService{
		userDAO: dao.NewUserDAO(),
	}
}

func (s *UserService) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	cacheKey := fmt.Sprintf(userCacheKey, id)

	// L1: 本地缓存
	if user := s.getUserFromLocalCache(cacheKey); user != nil {
		return user, nil
	}

	// L2: Redis 缓存
	if user := s.getUserFromRedis(ctx, cacheKey); user != nil {
		return user, nil
	}

	// L3: 使用 singleflight 防止缓存击穿，从数据库获取
	result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
		// 双重检查 Redis
		if user := s.getUserFromRedis(ctx, cacheKey); user != nil {
			return user, nil
		}
		return s.loadUserFromDB(ctx, id, cacheKey)
	})

	if err != nil {
		return nil, err
	}
	return result.(*model.User), nil
}

// getUserFromLocalCache 从本地缓存获取用户
func (s *UserService) getUserFromLocalCache(cacheKey string) *model.User {
	lc := cache.GetLocalCache()
	if lc == nil {
		return nil
	}
	if val, ok := lc.Get(cacheKey); ok {
		if user, ok := val.(*model.User); ok {
			return user
		}
	}
	return nil
}

// getUserFromRedis 从 Redis 缓存获取用户
//
//nolint:dupl // 与 getPageFromRedis 结构相似但类型不同，保持类型安全
func (s *UserService) getUserFromRedis(ctx context.Context, cacheKey string) *model.User {
	if cache.RDB == nil {
		return nil
	}
	cached, err := cache.Get(ctx, cacheKey)
	if err != nil || cached == "" {
		return nil
	}
	var user model.User
	if err := sonic.UnmarshalString(cached, &user); err != nil {
		return nil
	}
	// 回填本地缓存
	if lc := cache.GetLocalCache(); lc != nil {
		lc.SetWithTTL(cacheKey, &user, 1, localCacheTTL)
	}
	return &user
}

// loadUserFromDB 从数据库加载用户并写入缓存
func (s *UserService) loadUserFromDB(ctx context.Context, id uint64, cacheKey string) (*model.User, error) {
	user, err := s.userDAO.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.cacheUser(ctx, cacheKey, user)
	return user, nil
}

// cacheUser 将用户写入缓存
func (s *UserService) cacheUser(ctx context.Context, cacheKey string, user *model.User) {
	// 写入 Redis 缓存
	if cache.RDB != nil {
		data, _ := sonic.MarshalString(user)
		if err := cache.Set(ctx, cacheKey, data, cacheTTL); err != nil {
			logger.WarnCtxf(ctx, "failed to cache user", "key", cacheKey, "error", err)
		}
	}
	// 写入本地缓存
	if lc := cache.GetLocalCache(); lc != nil {
		lc.SetWithTTL(cacheKey, user, 1, localCacheTTL)
	}
}

func (s *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	// 先从 Redis 获取
	if cache.RDB != nil {
		cached, err := cache.Get(ctx, userListCacheKey)
		if err == nil && cached != "" {
			var users []model.User
			if err := sonic.UnmarshalString(cached, &users); err == nil {
				return users, nil
			}
		}
	}

	// 从数据库获取
	users, err := s.userDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if cache.RDB != nil {
		data, _ := sonic.MarshalString(users)
		if err := cache.Set(ctx, userListCacheKey, data, cacheTTL); err != nil {
			logger.WarnCtxf(ctx, "failed to cache user list", "key", userListCacheKey, "error", err)
		}
	}

	return users, nil
}

// GetPage 分页获取用户列表（多级缓存）
func (s *UserService) GetPage(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	page := offset/limit + 1
	cacheKey := fmt.Sprintf(userPageCacheKey, page, limit)

	// L1: 本地缓存
	if result := s.getPageFromLocalCache(cacheKey); result != nil {
		return result.Users, result.Total, nil
	}

	// L2: Redis 缓存
	if result := s.getPageFromRedis(ctx, cacheKey); result != nil {
		return result.Users, result.Total, nil
	}

	// L3: 使用 singleflight 防止缓存击穿
	result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
		// 双重检查 Redis
		if r := s.getPageFromRedis(ctx, cacheKey); r != nil {
			return r, nil
		}
		return s.loadPageFromDB(ctx, offset, limit, cacheKey)
	})

	if err != nil {
		return nil, 0, err
	}

	r := result.(*model.UserListResult)
	return r.Users, r.Total, nil
}

// getPageFromLocalCache 从本地缓存获取分页结果
func (s *UserService) getPageFromLocalCache(cacheKey string) *model.UserListResult {
	lc := cache.GetLocalCache()
	if lc == nil {
		return nil
	}
	if val, ok := lc.Get(cacheKey); ok {
		if result, ok := val.(*model.UserListResult); ok {
			return result
		}
	}
	return nil
}

// getPageFromRedis 从 Redis 缓存获取分页结果
//
//nolint:dupl // 与 getUserFromRedis 结构相似但类型不同，保持类型安全
func (s *UserService) getPageFromRedis(ctx context.Context, cacheKey string) *model.UserListResult {
	if cache.RDB == nil {
		return nil
	}
	cached, err := cache.Get(ctx, cacheKey)
	if err != nil || cached == "" {
		return nil
	}
	var result model.UserListResult
	if err := sonic.UnmarshalString(cached, &result); err != nil {
		return nil
	}
	// 回填本地缓存
	if lc := cache.GetLocalCache(); lc != nil {
		lc.SetWithTTL(cacheKey, &result, 1, localCacheTTL)
	}
	return &result
}

// loadPageFromDB 从数据库加载分页数据并写入缓存
func (s *UserService) loadPageFromDB(ctx context.Context, offset, limit int, cacheKey string) (*model.UserListResult, error) {
	users, total, err := s.userDAO.GetPage(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	listResult := &model.UserListResult{
		Users: users,
		Total: total,
	}
	s.cachePageResult(ctx, cacheKey, listResult)
	return listResult, nil
}

// cachePageResult 将分页结果写入缓存
func (s *UserService) cachePageResult(ctx context.Context, cacheKey string, result *model.UserListResult) {
	// 写入 Redis 缓存
	if cache.RDB != nil {
		data, _ := sonic.MarshalString(result)
		if err := cache.Set(ctx, cacheKey, data, cacheTTL); err != nil {
			logger.WarnCtxf(ctx, "failed to cache user page", "key", cacheKey, "error", err)
		}
	}
	// 写入本地缓存
	if lc := cache.GetLocalCache(); lc != nil {
		lc.SetWithTTL(cacheKey, result, 1, localCacheTTL)
	}
}

// GetByIDs 批量获取用户（带缓存）
func (s *UserService) GetByIDs(ctx context.Context, ids []uint64) ([]model.User, error) {
	if len(ids) == 0 {
		return []model.User{}, nil // 返回空切片而非 nil，便于调用方判断
	}

	// 尝试从缓存批量获取
	users := make([]model.User, 0, len(ids))
	missingIDs := make([]uint64, 0)

	if cache.RDB != nil {
		for _, id := range ids {
			cacheKey := fmt.Sprintf(userCacheKey, id)
			cached, err := cache.Get(ctx, cacheKey)
			if err == nil && cached != "" {
				var user model.User
				if err := sonic.UnmarshalString(cached, &user); err == nil {
					users = append(users, user)
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}
	} else {
		missingIDs = ids
	}

	// 批量查询缺失的
	if len(missingIDs) > 0 {
		dbUsers, err := s.userDAO.GetByIDs(ctx, missingIDs)
		if err != nil {
			return nil, err
		}

		// 写入缓存
		if cache.RDB != nil {
			for _, user := range dbUsers {
				cacheKey := fmt.Sprintf(userCacheKey, user.ID)
				data, _ := sonic.MarshalString(user)
				if err := cache.Set(ctx, cacheKey, data, cacheTTL); err != nil {
					logger.WarnCtxf(ctx, "failed to cache user in batch", "key", cacheKey, "error", err)
				}
			}
		}

		users = append(users, dbUsers...)
	}

	return users, nil
}

func (s *UserService) Create(ctx context.Context, user *model.User) error {
	if err := s.userDAO.Create(ctx, user); err != nil {
		return err
	}

	// 清除分页缓存
	s.invalidatePageCache(ctx)

	return nil
}

func (s *UserService) Update(ctx context.Context, user *model.User) error {
	if err := s.userDAO.Update(ctx, user); err != nil {
		return err
	}

	// 清除相关缓存
	s.invalidateUserCache(ctx, user.ID)

	return nil
}

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	if err := s.userDAO.Delete(ctx, id); err != nil {
		return err
	}

	// 清除相关缓存
	s.invalidateUserCache(ctx, id)
	s.invalidatePageCache(ctx)

	return nil
}

// invalidateUserCache 清除用户缓存
func (s *UserService) invalidateUserCache(ctx context.Context, id uint64) {
	if cache.RDB != nil {
		cacheKey := fmt.Sprintf(userCacheKey, id)
		_ = cache.Del(ctx, cacheKey, userListCacheKey)
	}
}

// invalidatePageCache 清除分页缓存
func (s *UserService) invalidatePageCache(ctx context.Context) {
	if cache.RDB == nil {
		return
	}

	// 使用 SCAN 替代 KEYS，避免阻塞 Redis
	pattern := "users:page:*"
	var cursor uint64
	var keys []string
	const scanCount = 100 // 每次扫描的数量

	for {
		var err error
		var batch []string
		batch, cursor, err = cache.RDB.Scan(ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			logger.WarnCtxf(ctx, "failed to scan page cache keys", "error", err)
			break
		}
		keys = append(keys, batch...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := cache.Del(ctx, keys...); err != nil {
			logger.WarnCtxf(ctx, "failed to delete page cache keys", "count", len(keys), "error", err)
		}
	}
	if err := cache.Del(ctx, userListCacheKey, userCountCacheKey); err != nil {
		logger.WarnCtxf(ctx, "failed to delete list/count cache", "error", err)
	}
}
