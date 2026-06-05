// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sse

import (
	"context"

	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EventsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewEventsLogic SSE 事件流
func NewEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EventsLogic {
	return &EventsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EventsLogic) Events(client chan<- *types.Response) error {
	// todo: add your logic here and delete this line

	return nil
}
