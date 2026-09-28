package userclient

import (
	"context"
	"demo-shop-back/src/contracts"
	"demo-shop-back/src/infra/cache"
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
	perm  v1_userv1.PermissionServiceClient
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
		perm:  v1_userv1.NewPermissionServiceClient(client.Conn()),
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

	resp, err := c.perm.ListPermCodesByUserId(ctx, &v1_userv1.ListPermCodesByUserIdReq{UserId: userId})
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

	resp, err := c.perm.ListPermCodesByApi(ctx, &v1_userv1.ListPermCodesByApiReq{
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
