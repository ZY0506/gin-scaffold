package application

import (
	"context"
	"time"

	domain "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/domain"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

type BlacklistService struct {
	repo domain.BlacklistRepository
}

func NewBlacklistService(repo domain.BlacklistRepository) *BlacklistService {
	return &BlacklistService{repo: repo}
}

// Add 添加黑名单记录
func (s *BlacklistService) Add(ctx context.Context, req *AddBlacklistReq, operatorID uint) error {
	if req.UserID == nil && req.IP == "" {
		return errors.New(errors.ErrBadRequest, "用户ID和IP不能同时为空")
	}

	bl := &domain.Blacklist{
		UserID:    req.UserID,
		IP:        req.IP,
		Reason:    req.Reason,
		BlockedBy: operatorID,
		BlockedAt: time.Now(),
		IsActive:  true,
	}

	if req.ExpiredAt != "" {
		t, err := time.Parse(time.DateTime, req.ExpiredAt)
		if err != nil {
			return errors.New(errors.ErrBadRequest, "到期时间格式错误，请使用 YYYY-MM-DD HH:mm:ss")
		}
		bl.ExpiredAt = &t
	}

	return s.repo.Create(ctx, bl)
}

// List 分页查询黑名单列表
func (s *BlacklistService) List(ctx context.Context, req *BlacklistListReq) ([]BlacklistItemResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	list, total, err := s.repo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrDB, "查询黑名单失败")
	}

	resp := make([]BlacklistItemResp, len(list))
	for i, bl := range list {
		item := BlacklistItemResp{
			ID:        bl.ID,
			UserID:    bl.UserID,
			IP:        bl.IP,
			Reason:    bl.Reason,
			BlockedBy: bl.BlockedBy,
			BlockedAt: bl.BlockedAt.Format(time.DateTime),
			IsActive:  bl.IsActive,
		}
		if bl.ExpiredAt != nil {
			item.ExpiredAt = bl.ExpiredAt.Format(time.DateTime)
		}
		resp[i] = item
	}

	return resp, total, nil
}

// Deactivate 解封（软解除）
func (s *BlacklistService) Deactivate(ctx context.Context, id uint) error {
	return s.repo.Deactivate(ctx, id)
}
