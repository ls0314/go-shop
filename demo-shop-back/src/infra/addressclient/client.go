package addressclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ErrUnavailable 客户端未建连(etcd 连不上 / 服务未注册)时的统一错误。
//
// 文案是契约:HTTP 层据此回 503。与 productclient / couponclient /
// tradeclient 的同名变量同一约定。
var ErrUnavailable = errors.New("user-service 不可用")

// callTimeout 单次地址域 RPC 的超时。
//
// 地址操作是单表本地事务,比下单那种 Saga 编排短得多 ——
// 3s 足够,而给太长会让"user-service 卡住"传导成地址簿打不开。
//
// 注意**下单取快照**那条路径在 trade 侧另有超时(见 tradeclient),
// 它需要更短:下单链路不能因为地址服务慢而整体卡住。
const callTimeout = 3 * time.Second

// AddressClient 地址域(user-service 的 AddressService)客户端。
type AddressClient struct {
	address v1_userv1.AddressServiceClient
	conn    *grpc.ClientConn
}

// NewAddressClient 建连 user-service(etcd 服务发现)。
//
// 与 UserClient 是**两条独立连接**:UserService 与 AddressService 虽然由
// 同一个进程提供、同一个 etcd key 发现,但分开建连能让"地址域不可用"
// 与"身份域不可用"在日志与超时配置上各自可辨。若将来连接数成为问题,
// 可以合并成一条 —— 那时要把两个客户端的超时统一,不能各留一份。
func NewAddressClient(etcdHosts []string, etcdKey string) (*AddressClient, error) {
	client, err := zrpc.NewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{Hosts: etcdHosts, Key: etcdKey},
	})
	if err != nil {
		return nil, err
	}
	conn := client.Conn()
	return &AddressClient{
		address: v1_userv1.NewAddressServiceClient(conn),
		conn:    conn,
	}, nil
}

func (c *AddressClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// RestoreError 把服务端 error_msg 还原成本地哨兵错误。
//
// **必须还原而不是 errors.New(msg) 了事**:HTTP 层要能区分
// "地址不存在"(404 语义)与"越权修改"(403 语义),而它们都是
// error_msg —— 靠文案比对太脆,靠 errors.Is 才稳。
// 与 couponclient.RestoreError 同一手法。
func RestoreError(errorMsg string) error {
	if errorMsg == "" {
		return nil
	}
	if err, ok := serverErrMap[errorMsg]; ok {
		return err
	}
	if errorMsg == ErrUnavailable.Error() {
		return ErrUnavailable
	}
	return errors.New(errorMsg)
}

// serverErrMap 服务端文案 → 本地哨兵错误。
// 地址域文案两侧逐字一致(都与单体 error_info.go 对齐),故是纯粹的还原。
var serverErrMap = map[string]error{
	"手机号格式错误":       model.PhoneMalformed,
	"地址不存在":         model.AddressNotExist,
	"地址列表为空":        model.AddressListIsNull,
	"地址数量已达上限":      model.AddressNumsIsFull,
	"用户越权修改地址":      model.UserNotSetAddress,
	"收货人姓名与手机号不能为空": model.ReceiverNotNull,
}

// Address 新增/更新地址的请求体。
//
// **不再借用 model.UserAddress**:那个实体是"demo_shop 里一行的形状"
// (带 is_deleted 之类由服务端决定的列),且它随本次迁移已成为旧实现的
// 残留物。请求体有自己的形状,两者混用会让"客户端能传哪些字段"
// 从代码上看不出来 —— 与 cart_item_handler 那次是同一个问题。
//
// 字段名与 JSON tag 必须与前端现有的请求体**逐字一致**:
// 这是已在用的 HTTP 契约,换实现不能改名。
type Address struct {
	UserId        int64  `json:"user_id"`
	ReceiverName  string `json:"receiver_name"`
	ReceiverPhone string `json:"receiver_phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	DetailAddress string `json:"detail_address"`
	PostalCode    string `json:"postal_code"`
	IsDefault     bool   `json:"is_default"`
	AddressTag    string `json:"address_tag"`
}

// ============================================================
// 6 个接口
// ============================================================

// CreateAddress 新增地址,返回新地址 ID
func (c *AddressClient) CreateAddress(address *Address) (int64, error) {
	if c == nil {
		return 0, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.address.CreateAddress(ctx, &v1_userv1.CreateAddressReq{
		UserId:        address.UserId,
		ReceiverName:  address.ReceiverName,
		ReceiverPhone: address.ReceiverPhone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		DetailAddress: address.DetailAddress,
		PostalCode:    address.PostalCode,
		IsDefault:     address.IsDefault,
		AddressTag:    address.AddressTag,
	})
	if err != nil {
		return 0, err
	}
	if resp.ErrorMsg != "" {
		return 0, RestoreError(resp.ErrorMsg)
	}
	return resp.AddressId, nil
}

// GetAddressList 用户地址列表
func (c *AddressClient) GetAddressList(userId int64) ([]*response.AddressResp, error) {
	if c == nil {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.address.ListAddresses(ctx, &v1_userv1.ListAddressesReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	if resp.ErrorMsg != "" {
		return nil, RestoreError(resp.ErrorMsg)
	}

	// 空列表是正常状态(新用户还没填地址),返回空切片而不是错误
	out := make([]*response.AddressResp, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, toModelAddress(it))
	}
	return out, nil
}

// GetAddress 单个地址详情
func (c *AddressClient) GetAddress(userId, addressId int64) (*response.AddressResp, error) {
	if c == nil {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.address.GetAddress(ctx, &v1_userv1.GetAddressReq{
		UserId:    userId,
		AddressId: addressId,
	})
	if err != nil {
		return nil, err
	}
	if resp.ErrorMsg != "" {
		return nil, RestoreError(resp.ErrorMsg)
	}
	return toModelAddress(resp.Address), nil
}

// UpdateAddress 局部更新地址
//
// updates 是前端传来的原始 map(只含要改的字段)。必须先转成
// []FieldUpdate 再发 —— proto3 的标量区分不了"没传"与"传了零值",
// 而地址的部分更新必须能区分:否则一次只改手机号的请求会把 is_default
// 一起抹成 false,用户就莫名失去了默认地址。
func (c *AddressClient) UpdateAddress(userId, addressId int64, updates map[string]interface{}) (*response.AddressResp, error) {
	if c == nil {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	fields, err := mapToFieldUpdates(updates)
	if err != nil {
		return nil, err
	}

	resp, err := c.address.UpdateAddress(ctx, &v1_userv1.UpdateAddressReq{
		UserId:    userId,
		AddressId: addressId,
		Updates:   fields,
	})
	if err != nil {
		return nil, err
	}
	if resp.ErrorMsg != "" {
		return nil, RestoreError(resp.ErrorMsg)
	}
	return toModelAddress(resp.Address), nil
}

// DeleteAddress 删除地址(软删除)
func (c *AddressClient) DeleteAddress(userId, addressId int64) error {
	if c == nil {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.address.DeleteAddress(ctx, &v1_userv1.DeleteAddressReq{
		UserId:    userId,
		AddressId: addressId,
	})
	if err != nil {
		return err
	}
	return RestoreError(resp.ErrorMsg)
}

// SetDefaultAddress 设为默认地址
func (c *AddressClient) SetDefaultAddress(userId, addressId int64) error {
	if c == nil {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.address.SetDefaultAddress(ctx, &v1_userv1.SetDefaultAddressReq{
		UserId:    userId,
		AddressId: addressId,
	})
	if err != nil {
		return err
	}
	return RestoreError(resp.ErrorMsg)
}

// GetAddressSnapshot 取下单用的地址快照。
//
// 单独一个方法而不是复用 GetAddress:两者字段集不同(快照不含管理属性),
// 而**失败语义更不同** —— 下单时"地址不存在"与"user-service 不可用"
// 必须能分开,否则会给用户一个"地址填错了"的误导提示。
// 本方法把前者还原成 model.AddressNotExist,后者原样返回 ErrUnavailable。
func (c *AddressClient) GetAddressSnapshot(userId, addressId int64) (*response.AddressResp, error) {
	if c == nil {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	resp, err := c.address.GetAddressSnapshot(ctx, &v1_userv1.GetAddressSnapshotReq{
		UserId:    userId,
		AddressId: addressId,
	})
	if err != nil {
		return nil, err
	}
	if resp.ErrorMsg != "" {
		return nil, RestoreError(resp.ErrorMsg)
	}
	// 快照只有发货需要的七个字段,管理属性(userId/isDefault/addressTag/
	// createdAt)在快照里没有,保持零值 —— 不要伪造它们
	s := resp.Snapshot
	return &response.AddressResp{
		ReceiverName:  s.GetReceiverName(),
		ReceiverPhone: s.GetReceiverPhone(),
		Province:      s.GetProvince(),
		City:          s.GetCity(),
		District:      s.GetDistrict(),
		DetailAddress: s.GetDetailAddress(),
		PostalCode:    s.GetPostalCode(),
	}, nil
}

// ============================================================
// 转换
// ============================================================

// toModelAddress proto → 单体 response
//
// 保留单体原有的 response.AddressResp 而不直接把 proto 交给 handler:
// HTTP 响应的 JSON 字段名是对前端的契约,已在用的字段不能因为换实现而改名。
// 注意 CreatedAt 在单体里是 **string**(前端展示用),故这里做格式化。
func toModelAddress(a *v1_userv1.Address) *response.AddressResp {
	if a == nil {
		return nil
	}
	return &response.AddressResp{
		AddressId:     a.AddressId,
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		DetailAddress: a.DetailAddress,
		PostalCode:    a.PostalCode,
		IsDefault:     a.IsDefault,
		AddressTag:    a.AddressTag,
		CreatedAt:     formatTime(a.CreatedAt),
	}
}

// formatTime 时间格式化。
//
// 用与单体 `time.Time.String()` **相同**的布局(即 Go 的默认格式),
// 而不是 RFC3339:前端的地址列表直接展示这个字符串,换格式会让页面上的
// 时间看起来变了。这是"照搬既有行为"而非"选一个更好的格式"。
func formatTime(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().String()
}

// mapToFieldUpdates 把前端传来的更新 map 转成 proto 的 FieldUpdate 列表。
//
// 只接受**已知字段**:未知键直接报错而不是忽略。忽略会让"前端拼错了字段名"
// 表现成"更新成功了但字段没变",那是最难排查的一类问题。
//
// 值的类型必须与 proto 的 oneof 分支匹配:string / bool / 数值。
// 前端的 JSON 数字会以 float64 落进 interface{},故数值统一转 int64 ——
// 地址的数值字段只有 address_id 与 user_id,而那两个不该由前端更新。
func mapToFieldUpdates(updates map[string]interface{}) ([]*v1_userv1.FieldUpdate, error) {
	out := make([]*v1_userv1.FieldUpdate, 0, len(updates))
	for field, raw := range updates {
		if !editableFields[field] {
			return nil, fmt.Errorf("不支持更新的字段: %s", field)
		}

		var value *v1_userv1.FieldValue
		switch v := raw.(type) {
		case string:
			value = &v1_userv1.FieldValue{
				Value: &v1_userv1.FieldValue_StringValue{StringValue: v},
			}
		case bool:
			value = &v1_userv1.FieldValue{
				Value: &v1_userv1.FieldValue_BoolValue{BoolValue: v},
			}
		case float64:
			// encoding/json 把所有数字解成 float64
			value = &v1_userv1.FieldValue{
				Value: &v1_userv1.FieldValue_Int64Value{Int64Value: int64(v)},
			}
		case nil:
			// null:跳过而不是报错。前端清空一个可选字段(如 postal_code)
			// 时会传 null,而"清空"可以走空串表达 —— 让它在下面按空串处理
			value = &v1_userv1.FieldValue{
				Value: &v1_userv1.FieldValue_StringValue{StringValue: ""},
			}
		default:
			return nil, fmt.Errorf("字段 %s 的值类型不支持: %T", field, raw)
		}

		out = append(out, &v1_userv1.FieldUpdate{Field: field, Value: value})
	}
	return out, nil
}

// editableFields 前端可更新的字段(与 proto 的 JSON 名一致)。
//
// user_id / address_id 不在其中:它们是**所有权**,不是业务字段。
// 允许改 user_id 等于允许把地址转给别的用户。
var editableFields = map[string]bool{
	"receiver_name":  true,
	"receiver_phone": true,
	"province":       true,
	"city":           true,
	"district":       true,
	"detail_address": true,
	"postal_code":    true,
	"is_default":     true,
	"address_tag":    true,
}
