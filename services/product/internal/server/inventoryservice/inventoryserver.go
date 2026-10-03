package inventoryservice

import (
	"context"
	"demo-shop/api/gen/product/v1"
	"demo-shop/services/product/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type InventoryServiceServer struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	v1_productv1.UnimplementedInventoryServiceServer
}

func NewInventoryServiceServer(svcCtx *svc.ServiceContext) *InventoryServiceServer {
	return &InventoryServiceServer{
		Logger: logx.WithContext(context.Background()),
		ctx:    context.Background(),
		svcCtx: svcCtx,
	}
}
