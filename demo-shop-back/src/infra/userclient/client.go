package userclient

import (
	"context"
	"demo-shop-back/src/contracts"
	"demo-shop-back/src/infra/cache"
	"demo-shop-back/src/model"
	"demo-shop/api/gen/user/v1"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

const (
	userPermKeyFmt  = "user:perm:%d"
	apiPermKeyFmt   = "api:perm:v%s:%s:%s"
	permVersionKey  = "api:perm:version"
	permCacheTTL    = 30 * time.Minute
	permCallTimeout = 2 * time.Second
)

type PermCodesClient struct {
	rbac  v1_userv1.RBACServiceClient
	cache *cache.RedisService
	conn  *grpc.ClientConn
}

var _ contracts.PermCodesSource = (*PermCodesClient)(nil)

func NewPermCodesClient(etcdHosts []string, etcdKey string, cch *cache.RedisService) (*PermCodesClient, error) {
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: etcdHosts, Key: etcdKey},
	})
	if err != nil {
		return nil, err
	}
	return &PermCodesClient{
		rbac:  v1_userv1.NewRBACServiceClient(client.Conn()),
		cache: cch,
		conn:  client.Conn(),
	}, nil
}

func (c *PermCodesClient) Close() error { return c.conn.Close() }

// GetPermCodesByUserId 取用户全部权限码:缓存优先,miss 时经 RPC 取并回写。
func (c *PermCodesClient) GetPermCodesByUserId(userId int64) ([]string, error) {
	if c == nil {
		return nil, errors.New("user-service 不可用")
	}

	key := fmt.Sprintf(userPermKeyFmt, userId)
	if c.cache != nil {
		var codes []string
		hit, err := c.cache.GetJSON(context.Background(), key, &codes)
		if err == nil && hit {
			return codes, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListPermCodesByUserId(ctx, &v1_userv1.ListPermCodesByUserIdReq{UserId: userId})
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		_ = c.cache.SetJSON(context.Background(), key, resp.PermCodes, permCacheTTL)
	}
	return resp.PermCodes, nil
}

// permVersion 读权限版本号,读不到按 0 处理。
// 读不到不等于失败:版本号缺席时应退回"无版本"缓存,而不是让整个判权失败。
func (c *PermCodesClient) permVersion() string {
	if c.cache == nil {
		return "0"
	}
	v, err := c.cache.Get(context.Background(), permVersionKey)
	if err == nil {
		return v
	}
	return "0"
}

// GetPermCodesByApi 取接口要求的权限码:按权限版本号缓存,miss 时经 RPC 取。
func (c *PermCodesClient) GetPermCodesByApi(path, method string) ([]string, error) {
	if c == nil {
		return nil, errors.New("user-service 不可用")
	}

	key := fmt.Sprintf(apiPermKeyFmt, c.permVersion(), method, path)
	if c.cache != nil {
		var codes []string
		hit, err := c.cache.GetJSON(context.Background(), key, &codes)
		if err == nil && hit {
			return codes, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListPermCodesByApi(ctx, &v1_userv1.ListPermCodesByApiReq{
		ApiPath:       path,
		RequestMethod: method,
	})
	if err != nil {
		return nil, err
	}

	// 空结果不缓存:避免"未配权限码"的接口在配置生效前一直命中空缓存
	if c.cache != nil && len(resp.PermCodes) > 0 {
		_ = c.cache.SetJSON(context.Background(), key, resp.PermCodes, permCacheTTL)
	}
	return resp.PermCodes, nil
}

// ============ 权限点 CRUD ============

// GetPermission 按 ID 取权限点
func (c *PermCodesClient) GetPermission(id int64) (*model.SysPermission, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetPermission(ctx, &v1_userv1.GetPermissionReq{PermissionId: id})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelPermission(resp.Permission), "", nil
}

// ListPermissions 分页取权限点
func (c *PermCodesClient) ListPermissions(page, pageSize int, permType string) ([]model.SysPermission, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListPermissions(ctx, &v1_userv1.ListPermissionsReq{
		Page:           int32(page),
		PageSize:       int32(pageSize),
		PermissionType: permType,
	})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	perms := make([]model.SysPermission, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		perms = append(perms, *toModelPermission(item))
	}
	return perms, resp.Total, "", nil
}

// CreatePermission 创建权限点
func (c *PermCodesClient) CreatePermission(perm *model.SysPermission) (*model.SysPermission, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.CreatePermission(ctx, &v1_userv1.CreatePermissionReq{
		Permission: toProtoPermission(perm),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelPermission(resp.Permission), "", nil
}

// UpdatePermission 局部更新权限点,updates 的 key 为 JSON 字段名
func (c *PermCodesClient) UpdatePermission(id int64, updates map[string]interface{}) (*model.SysPermission, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	fields := make([]*v1_userv1.FieldUpdate, 0, len(updates))
	for k, v := range updates {
		fv, ok := toProtoFieldValue(v)
		if !ok {
			continue
		}
		fields = append(fields, &v1_userv1.FieldUpdate{Field: k, Value: fv})
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.UpdatePermission(ctx, &v1_userv1.UpdatePermissionReq{
		PermissionId: id,
		Updates:      fields,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelPermission(resp.Permission), "", nil
}

// DeletePermission 删除权限点
func (c *PermCodesClient) DeletePermission(id int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.DeletePermission(ctx, &v1_userv1.DeletePermissionReq{PermissionId: id})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ============ 类型转换 ============

// toModelPermission proto -> 单体 model。created_at 不在 proto 中,取零值。
func toModelPermission(p *v1_userv1.Permission) *model.SysPermission {
	if p == nil {
		return nil
	}
	return &model.SysPermission{
		PermissionID:   p.PermissionId,
		PermissionCode: p.PermissionCode,
		PermissionName: p.PermissionName,
		PermissionType: p.PermissionType,
		RequestMethod:  p.RequestMethod,
		ApiPath:        p.ApiPath,
		Description:    p.Description,
		IsSystem:       p.IsSystem,
	}
}

// toProtoPermission 单体 model -> proto
func toProtoPermission(p *model.SysPermission) *v1_userv1.Permission {
	if p == nil {
		return nil
	}
	return &v1_userv1.Permission{
		PermissionId:   p.PermissionID,
		PermissionCode: p.PermissionCode,
		PermissionName: p.PermissionName,
		PermissionType: p.PermissionType,
		RequestMethod:  p.RequestMethod,
		ApiPath:        p.ApiPath,
		Description:    p.Description,
		IsSystem:       p.IsSystem,
	}
}

// toProtoFieldValue 把 JSON 反序列化出的值装进 oneof。
// 只支持 string/int64/bool —— sys_permission 的可更新字段只有这三类。
func toProtoFieldValue(v interface{}) (*v1_userv1.FieldValue, bool) {
	switch x := v.(type) {
	case string:
		return &v1_userv1.FieldValue{Value: &v1_userv1.FieldValue_StringValue{StringValue: x}}, true
	case bool:
		return &v1_userv1.FieldValue{Value: &v1_userv1.FieldValue_BoolValue{BoolValue: x}}, true
	case float64:
		// JSON 数字默认解析为 float64
		return &v1_userv1.FieldValue{Value: &v1_userv1.FieldValue_Int64Value{Int64Value: int64(x)}}, true
	}
	return nil, false
}
