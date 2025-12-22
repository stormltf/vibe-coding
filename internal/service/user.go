package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/test-tt/internal/dao"
	"github.com/test-tt/internal/model"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/logger"
	"golang.org/x/sync/singleflight"
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

	// L1: 本地缓存（最快）
	if lc := cache.GetLocalCache(); lc != nil {
		if val, ok := lc.Get(cacheKey); ok {
			if user, ok := val.(*model.User); ok {
				return user, nil
			}
		}
	}

	// L2: Redis 缓存
	if cache.RDB != nil {
		cached, err := cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var user model.User
			if err := sonic.UnmarshalString(cached, &user); err == nil {
				// 回填本地缓存
				if lc := cache.GetLocalCache(); lc != nil {
					lc.SetWithTTL(cacheKey, &user, 1, localCacheTTL)
				}
				return &user, nil
			}
		}
	}

	// 使用 singleflight 防止缓存击穿
	result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
		// 双重检查 Redis
		if cache.RDB != nil {
			cached, err := cache.Get(ctx, cacheKey)
			if err == nil && cached != "" {
				var user model.User
				if err := sonic.UnmarshalString(cached, &user); err == nil {
					return &user, nil
				}
			}
		}

		// L3: 从数据库获取
		user, err := s.userDAO.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}

		// 写入 Redis 缓存
		if cache.RDB != nil {
			data, _ := sonic.MarshalString(user)
			_ = cache.Set(ctx, cacheKey, data, cacheTTL)
		}

		// 写入本地缓存
		if lc := cache.GetLocalCache(); lc != nil {
			lc.SetWithTTL(cacheKey, user, 1, localCacheTTL)
		}

		return user, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(*model.User), nil
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
		_ = cache.Set(ctx, userListCacheKey, data, cacheTTL)
	}

	return users, nil
}

// GetPage 分页获取用户列表（多级缓存）
func (s *UserService) GetPage(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	page := offset/limit + 1
	cacheKey := fmt.Sprintf(userPageCacheKey, page, limit)

	// L1: 本地缓存
	if lc := cache.GetLocalCache(); lc != nil {
		if val, ok := lc.Get(cacheKey); ok {
			if result, ok := val.(*model.UserListResult); ok {
				return result.Users, result.Total, nil
			}
		}
	}

	// L2: Redis 缓存
	if cache.RDB != nil {
		cached, err := cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var result model.UserListResult
			if err := sonic.UnmarshalString(cached, &result); err == nil {
				// 回填本地缓存
				if lc := cache.GetLocalCache(); lc != nil {
					lc.SetWithTTL(cacheKey, &result, 1, localCacheTTL)
				}
				return result.Users, result.Total, nil
			}
		}
	}

	// 使用 singleflight 防止缓存击穿
	result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
		// 双重检查 Redis
		if cache.RDB != nil {
			cached, err := cache.Get(ctx, cacheKey)
			if err == nil && cached != "" {
				var result model.UserListResult
				if err := sonic.UnmarshalString(cached, &result); err == nil {
					return &result, nil
				}
			}
		}

		// L3: 从数据库获取
		users, total, err := s.userDAO.GetPage(ctx, offset, limit)
		if err != nil {
			return nil, err
		}

		listResult := &model.UserListResult{
			Users: users,
			Total: total,
		}

		// 写入 Redis 缓存
		if cache.RDB != nil {
			data, _ := sonic.MarshalString(listResult)
			_ = cache.Set(ctx, cacheKey, data, cacheTTL)
		}

		// 写入本地缓存
		if lc := cache.GetLocalCache(); lc != nil {
			lc.SetWithTTL(cacheKey, listResult, 1, localCacheTTL)
		}

		return listResult, nil
	})

	if err != nil {
		return nil, 0, err
	}

	r := result.(*model.UserListResult)
	return r.Users, r.Total, nil
}

// GetByIDs 批量获取用户（带缓存）
func (s *UserService) GetByIDs(ctx context.Context, ids []uint64) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
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
				_ = cache.Set(ctx, cacheKey, data, cacheTTL)
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

	// 使用 pattern 删除所有分页缓存
	pattern := "users:page:*"
	keys, err := cache.RDB.Keys(ctx, pattern).Result()
	if err != nil {
		logger.WarnCtxf(ctx, "failed to get page cache keys", "error", err)
		return
	}

	if len(keys) > 0 {
		_ = cache.Del(ctx, keys...)
	}
	_ = cache.Del(ctx, userListCacheKey, userCountCacheKey)
}
