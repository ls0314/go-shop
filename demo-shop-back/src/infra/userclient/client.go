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
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
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

// ============ 角色 CRUD ============

// GetRole 按 ID 取角色
func (c *PermCodesClient) GetRole(id int64) (*model.SysRole, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetRole(ctx, &v1_userv1.GetRoleReq{RoleId: id})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelRole(resp.Role), "", nil
}

// ListRoles 分页取角色
func (c *PermCodesClient) ListRoles(page, pageSize int, roleType string) ([]model.SysRole, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListRoles(ctx, &v1_userv1.ListRolesReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		RoleType: roleType,
	})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	roles := make([]model.SysRole, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		roles = append(roles, *toModelRole(item))
	}
	return roles, resp.Total, "", nil
}

// CreateRole 创建角色
func (c *PermCodesClient) CreateRole(role *model.SysRole) (*model.SysRole, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.CreateRole(ctx, &v1_userv1.CreateRoleReq{
		Role: toProtoRole(role),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelRole(resp.Role), "", nil
}

// UpdateRole 局部更新角色,updates 的 key 为 JSON 字段名
func (c *PermCodesClient) UpdateRole(id int64, updates map[string]interface{}) (*model.SysRole, string, error) {
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

	resp, err := c.rbac.UpdateRole(ctx, &v1_userv1.UpdateRoleReq{
		RoleId:  id,
		Updates: fields,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelRole(resp.Role), "", nil
}

// DeleteRole 删除角色
func (c *PermCodesClient) DeleteRole(id int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.DeleteRole(ctx, &v1_userv1.DeleteRoleReq{RoleId: id})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ============ 菜单 CRUD ============

// GetMenu 按 ID 取菜单
func (c *PermCodesClient) GetMenu(id int64) (*model.SysMenu, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetMenu(ctx, &v1_userv1.GetMenuReq{MenuId: id})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelMenu(resp.Menu), "", nil
}

// ListMenus 分页取菜单,返回扁平列表(children 为空)
func (c *PermCodesClient) ListMenus(page, pageSize int, menuType string) ([]model.SysMenu, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListMenus(ctx, &v1_userv1.ListMenusReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		MenuType: menuType,
	})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	menus := make([]model.SysMenu, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		menus = append(menus, *toModelMenu(item))
	}
	return menus, resp.Total, "", nil
}

// CreateMenu 创建菜单
func (c *PermCodesClient) CreateMenu(menu *model.SysMenu) (*model.SysMenu, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.CreateMenu(ctx, &v1_userv1.CreateMenuReq{
		Menu: toProtoMenu(menu),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelMenu(resp.Menu), "", nil
}

// UpdateMenu 局部更新菜单,updates 的 key 为 JSON 字段名
func (c *PermCodesClient) UpdateMenu(id int64, updates map[string]interface{}) (*model.SysMenu, string, error) {
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

	resp, err := c.rbac.UpdateMenu(ctx, &v1_userv1.UpdateMenuReq{
		MenuId:  id,
		Updates: fields,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelMenu(resp.Menu), "", nil
}

// DeleteMenu 删除菜单
func (c *PermCodesClient) DeleteMenu(id int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.DeleteMenu(ctx, &v1_userv1.DeleteMenuReq{MenuId: id})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// GetMenuTreeByUserId 取指定用户的菜单树(用户 → 角色 → 菜单并集)
func (c *PermCodesClient) GetMenuTreeByUserId(userId int64) ([]*model.SysMenu, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetMenuTreeByUserId(ctx, &v1_userv1.GetMenuTreeByUserIdReq{UserId: userId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelMenuTree(resp.Items), "", nil
}

// GetMenuTreeByRoleId 取指定角色的菜单树。
// 当前前端由菜单列表接口自行建树,此方法暂无调用方,保留接口完整性。
func (c *PermCodesClient) GetMenuTreeByRoleId(roleId int64) ([]*model.SysMenu, error) {
	if c == nil {
		return nil, errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetMenuTreeByRoleId(ctx, &v1_userv1.GetMenuTreeByRoleIdReq{RoleId: roleId})
	if err != nil {
		return nil, err
	}
	if resp.ErrorMsg != "" {
		return nil, errors.New(resp.ErrorMsg)
	}
	return toModelMenuTree(resp.Items), nil
}

// ============ 部门 CRUD ============

// GetDept 按 ID 取部门
func (c *PermCodesClient) GetDept(id int64) (*model.SysDept, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetDept(ctx, &v1_userv1.GetDeptReq{DeptId: id})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelDept(resp.Dept), "", nil
}

// ListDepts 分页取部门,返回扁平列表(children 为空)
func (c *PermCodesClient) ListDepts(page, pageSize int, deptType string) ([]model.SysDept, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListDepts(ctx, &v1_userv1.ListDeptsReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		DeptType: deptType,
	})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	depts := make([]model.SysDept, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		depts = append(depts, *toModelDept(item))
	}
	return depts, resp.Total, "", nil
}

// CreateDept 创建部门
func (c *PermCodesClient) CreateDept(dept *model.SysDept) (*model.SysDept, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.CreateDept(ctx, &v1_userv1.CreateDeptReq{
		Dept: toProtoDept(dept),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelDept(resp.Dept), "", nil
}

// UpdateDept 局部更新部门,updates 的 key 为 JSON 字段名
func (c *PermCodesClient) UpdateDept(id int64, updates map[string]interface{}) (*model.SysDept, string, error) {
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

	resp, err := c.rbac.UpdateDept(ctx, &v1_userv1.UpdateDeptReq{
		DeptId:  id,
		Updates: fields,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelDept(resp.Dept), "", nil
}

// DeleteDept 删除部门
func (c *PermCodesClient) DeleteDept(id int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.DeleteDept(ctx, &v1_userv1.DeleteDeptReq{DeptId: id})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// GetDeptTreeByUserId 取指定用户所属部门的树
func (c *PermCodesClient) GetDeptTreeByUserId(userId int64) ([]*model.SysDept, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetDeptTreeByUserId(ctx, &v1_userv1.GetDeptTreeByUserIdReq{UserId: userId})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelDeptTree(resp.Items), "", nil
}

// ============ 数据权限 CRUD ============

// GetScope 按 ID 取数据权限
func (c *PermCodesClient) GetScope(id int64) (*model.SysScope, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.GetScope(ctx, &v1_userv1.GetScopeReq{ScopeId: id})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelScope(resp.Scope), "", nil
}

// ListScopes 分页取数据权限
func (c *PermCodesClient) ListScopes(page, pageSize int, resourceType string) ([]*model.SysScope, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListScopes(ctx, &v1_userv1.ListScopesReq{
		Page:         int32(page),
		PageSize:     int32(pageSize),
		ResourceType: resourceType,
	})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	scopes := make([]*model.SysScope, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		scopes = append(scopes, toModelScope(item))
	}
	return scopes, resp.Total, "", nil
}

// CreateScope 创建数据权限
func (c *PermCodesClient) CreateScope(scope *model.SysScope) (*model.SysScope, string, error) {
	if c == nil {
		return nil, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.CreateScope(ctx, &v1_userv1.CreateScopeReq{
		Scope: toProtoScope(scope),
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelScope(resp.Scope), "", nil
}

// UpdateScope 局部更新数据权限,updates 的 key 为 JSON 字段名
func (c *PermCodesClient) UpdateScope(id int64, updates map[string]interface{}) (*model.SysScope, string, error) {
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

	resp, err := c.rbac.UpdateScope(ctx, &v1_userv1.UpdateScopeReq{
		ScopeId: id,
		Updates: fields,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, resp.ErrorMsg, nil
	}
	return toModelScope(resp.Scope), "", nil
}

// DeleteScope 删除数据权限
func (c *PermCodesClient) DeleteScope(id int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.DeleteScope(ctx, &v1_userv1.DeleteScopeReq{ScopeId: id})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ============ 绑定关系 ============

// AssignRolePerms 全量替换角色的权限绑定
func (c *PermCodesClient) AssignRolePerms(roleId int64, permIds []int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.AssignRolePerms(ctx, &v1_userv1.AssignRolePermsReq{
		RoleId:  roleId,
		PermIds: permIds,
	})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ListRolePerms 查角色已绑定的权限
func (c *PermCodesClient) ListRolePerms(roleId int64) ([]*model.SysPermission, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListRolePerms(ctx, &v1_userv1.ListRolePermsReq{RoleId: roleId})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	perms := make([]*model.SysPermission, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		perms = append(perms, toModelPermission(item))
	}
	return perms, resp.Total, "", nil
}

// ClearRolePerms 清空角色的权限绑定
func (c *PermCodesClient) ClearRolePerms(roleId int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ClearRolePerms(ctx, &v1_userv1.ClearRolePermsReq{RoleId: roleId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// AssignRoleMenus 全量替换角色的菜单绑定
func (c *PermCodesClient) AssignRoleMenus(roleId int64, menuIds []int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.AssignRoleMenus(ctx, &v1_userv1.AssignRoleMenusReq{
		RoleId:  roleId,
		MenuIds: menuIds,
	})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ListRoleMenus 查角色已绑定的菜单
func (c *PermCodesClient) ListRoleMenus(roleId int64) ([]*model.SysMenu, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListRoleMenus(ctx, &v1_userv1.ListRoleMenusReq{RoleId: roleId})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}
	return toModelMenuTree(resp.Items), resp.Total, "", nil
}

// ClearRoleMenus 清空角色的菜单绑定
func (c *PermCodesClient) ClearRoleMenus(roleId int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ClearRoleMenus(ctx, &v1_userv1.ClearRoleMenusReq{RoleId: roleId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// AssignMenuPerms 全量替换菜单的权限绑定
func (c *PermCodesClient) AssignMenuPerms(menuId int64, permIds []int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.AssignMenuPerms(ctx, &v1_userv1.AssignMenuPermsReq{
		MenuId:  menuId,
		PermIds: permIds,
	})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ListMenuPerms 查菜单已绑定的权限
func (c *PermCodesClient) ListMenuPerms(menuId int64) ([]*model.SysPermission, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListMenuPerms(ctx, &v1_userv1.ListMenuPermsReq{MenuId: menuId})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	perms := make([]*model.SysPermission, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		perms = append(perms, toModelPermission(item))
	}
	return perms, resp.Total, "", nil
}

// ClearMenuPerms 清空菜单的权限绑定
func (c *PermCodesClient) ClearMenuPerms(menuId int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ClearMenuPerms(ctx, &v1_userv1.ClearMenuPermsReq{MenuId: menuId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// AssignUserRoles 全量替换用户的角色绑定
func (c *PermCodesClient) AssignUserRoles(userId int64, roleIds []int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.AssignUserRoles(ctx, &v1_userv1.AssignUserRolesReq{
		UserId:  userId,
		RoleIds: roleIds,
	})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ListUserRoles 查用户已绑定的角色
func (c *PermCodesClient) ListUserRoles(userId int64) ([]*model.SysRole, int64, string, error) {
	if c == nil {
		return nil, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListUserRoles(ctx, &v1_userv1.ListUserRolesReq{UserId: userId})
	if err != nil {
		return nil, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, resp.ErrorMsg, nil
	}

	roles := make([]*model.SysRole, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item == nil {
			continue
		}
		roles = append(roles, toModelRole(item))
	}
	return roles, resp.Total, "", nil
}

// ClearUserRoles 清空用户的角色绑定
func (c *PermCodesClient) ClearUserRoles(userId int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ClearUserRoles(ctx, &v1_userv1.ClearUserRolesReq{UserId: userId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// AssignUserDepts 全量替换用户的部门绑定,primaryIndex 是 deptIds 中的下标
func (c *PermCodesClient) AssignUserDepts(userId int64, deptIds []int64, primaryIndex int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.AssignUserDepts(ctx, &v1_userv1.AssignUserDeptsReq{
		UserId:       userId,
		DeptIds:      deptIds,
		PrimaryIndex: primaryIndex,
	})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
}

// ListUserDepts 查用户已绑定的部门,并返回主部门ID
func (c *PermCodesClient) ListUserDepts(userId int64) ([]*model.SysDept, int64, int64, string, error) {
	if c == nil {
		return nil, 0, 0, "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ListUserDepts(ctx, &v1_userv1.ListUserDeptsReq{UserId: userId})
	if err != nil {
		return nil, 0, 0, "", err
	}
	if resp.ErrorMsg != "" {
		return nil, 0, 0, resp.ErrorMsg, nil
	}
	return toModelDeptTree(resp.Items), resp.Total, resp.PrimaryDeptId, "", nil
}

// ClearUserDepts 清空用户的部门绑定
func (c *PermCodesClient) ClearUserDepts(userId int64) (string, error) {
	if c == nil {
		return "", errors.New("user-service 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), permCallTimeout)
	defer cancel()

	resp, err := c.rbac.ClearUserDepts(ctx, &v1_userv1.ClearUserDeptsReq{UserId: userId})
	if err != nil {
		return "", err
	}
	return resp.ErrorMsg, nil
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

// toModelRole proto -> 单体 model。
func toModelRole(r *v1_userv1.Role) *model.SysRole {
	if r == nil {
		return nil
	}
	return &model.SysRole{
		RoleId:      r.RoleId,
		RoleName:    r.RoleName,
		RoleType:    r.RoleType,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		IsDefault:   r.IsDefault,
		DataScope:   r.DataScope,
		CreatedAt:   r.GetCreatedAt().AsTime(),
	}
}

// toProtoRole 单体 model -> proto
func toProtoRole(r *model.SysRole) *v1_userv1.Role {
	if r == nil {
		return nil
	}
	return &v1_userv1.Role{
		RoleId:      r.RoleId,
		RoleName:    r.RoleName,
		RoleType:    r.RoleType,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		IsDefault:   r.IsDefault,
		DataScope:   r.DataScope,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
}

// toModelMenu proto -> 单体 model。
func toModelMenu(p *v1_userv1.Menu) *model.SysMenu {
	if p == nil {
		return nil
	}
	var meta datatypes.JSONMap
	if p.MetaInfo != nil {
		meta = datatypes.JSONMap(p.MetaInfo.AsMap())
	}
	return &model.SysMenu{
		MenuId:        p.MenuId,
		ParentId:      p.ParentId,
		MenuName:      p.MenuName,
		MenuType:      p.MenuType,
		Icon:          p.Icon,
		RoutePath:     p.RoutePath,
		ComponentPath: p.Component,
		IsVisible:     p.IsVisible,
		IsCache:       p.IsCache,
		SortOrder:     p.SortOrder,
		MetaInfo:      meta,
		CreatedAt:     p.GetCreatedAt().AsTime(),
	}
}

// toModelMenuTree 递归转换菜单树。
func toModelMenuTree(items []*v1_userv1.Menu) []*model.SysMenu {
	out := make([]*model.SysMenu, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		node := toModelMenu(item)
		node.Children = toModelMenuTree(item.Children)
		out = append(out, node)
	}
	return out
}

// toProtoMenu 单体 model -> proto,children 递归转换。
func toProtoMenu(m *model.SysMenu) *v1_userv1.Menu {
	if m == nil {
		return nil
	}
	children := make([]*v1_userv1.Menu, 0, len(m.Children))
	for _, c := range m.Children {
		children = append(children, toProtoMenu(c))
	}

	var meta *structpb.Struct
	if len(m.MetaInfo) > 0 {
		// meta_info 是展示用字段,类型无法转换时降级为空,不让整个请求失败
		meta, _ = structpb.NewStruct(m.MetaInfo)
	}

	return &v1_userv1.Menu{
		MenuId:    m.MenuId,
		ParentId:  m.ParentId,
		MenuName:  m.MenuName,
		MenuType:  m.MenuType,
		Icon:      m.Icon,
		RoutePath: m.RoutePath,
		Component: m.ComponentPath,
		IsVisible: m.IsVisible,
		IsCache:   m.IsCache,
		SortOrder: m.SortOrder,
		MetaInfo:  meta,
		CreatedAt: timestamppb.New(m.CreatedAt),
		Children:  children,
	}
}

// toModelDept proto -> 单体 model。
func toModelDept(p *v1_userv1.Dept) *model.SysDept {
	if p == nil {
		return nil
	}
	return &model.SysDept{
		DeptId:    p.DeptId,
		ParentId:  p.ParentId,
		DeptName:  p.DeptName,
		DeptType:  p.DeptType,
		LeaderId:  p.LeaderId,
		SortOrder: p.SortOrder,
		Status:    p.Status,
		CreatedAt: p.GetCreatedAt().AsTime(),
	}
}

// toModelDeptTree 递归转换部门树。
func toModelDeptTree(items []*v1_userv1.Dept) []*model.SysDept {
	out := make([]*model.SysDept, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		node := toModelDept(item)
		node.Children = toModelDeptTree(item.Children)
		out = append(out, node)
	}
	return out
}

// toProtoDept 单体 model -> proto,children 递归转换。
func toProtoDept(d *model.SysDept) *v1_userv1.Dept {
	if d == nil {
		return nil
	}
	children := make([]*v1_userv1.Dept, 0, len(d.Children))
	for _, c := range d.Children {
		children = append(children, toProtoDept(c))
	}
	return &v1_userv1.Dept{
		DeptId:    d.DeptId,
		ParentId:  d.ParentId,
		DeptName:  d.DeptName,
		DeptType:  d.DeptType,
		LeaderId:  d.LeaderId,
		SortOrder: d.SortOrder,
		Status:    d.Status,
		CreatedAt: timestamppb.New(d.CreatedAt),
		Children:  children,
	}
}

// toModelScope proto -> 单体 model。
func toModelScope(p *v1_userv1.Scope) *model.SysScope {
	if p == nil {
		return nil
	}
	return &model.SysScope{
		ScopeId:        p.ScopeId,
		RoleId:         p.RoleId,
		ResourceType:   p.ResourceType,
		FieldName:      p.FieldName,
		ConditionType:  p.ConditionType,
		ConditionValue: p.ConditionValue,
		Description:    p.Description,
		CreatedAt:      p.GetCreatedAt().AsTime(),
	}
}

// toProtoScope 单体 model -> proto
func toProtoScope(s *model.SysScope) *v1_userv1.Scope {
	if s == nil {
		return nil
	}
	return &v1_userv1.Scope{
		ScopeId:        s.ScopeId,
		RoleId:         s.RoleId,
		ResourceType:   s.ResourceType,
		FieldName:      s.FieldName,
		ConditionType:  s.ConditionType,
		ConditionValue: s.ConditionValue,
		Description:    s.Description,
		CreatedAt:      timestamppb.New(s.CreatedAt),
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
