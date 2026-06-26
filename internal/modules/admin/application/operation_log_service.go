package application

import (
	"context"

	adminDomain "github.com/ZY0506/gin-scaffold/internal/modules/admin/domain"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

// OperationLogService 操作日志服务
type OperationLogService struct {
	repo adminDomain.OperationLogRepository
}

func NewOperationLogService(repo adminDomain.OperationLogRepository) *OperationLogService {
	return &OperationLogService{repo: repo}
}

// Create 创建操作日志
func (s *OperationLogService) Create(ctx context.Context, log *adminDomain.OperationLog) error {
	return s.repo.Create(ctx, log)
}

// List 分页查询操作日志
func (s *OperationLogService) List(ctx context.Context, req *OperationLogListReq) ([]adminDomain.OperationLog, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	list, total, err := s.repo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrDB, "查询操作日志失败")
	}

	return list, total, nil
}
