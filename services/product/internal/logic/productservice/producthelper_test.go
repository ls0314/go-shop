package productservicelogic

import (
	"demo-shop/services/product/internal/model"
	"errors"
	"testing"

	"gorm.io/datatypes"
)

// 规格模板(两条规格:颜色 × 尺寸,笛卡尔积 = 4 个 SKU)
const validSpecTemplate = `[
  {"name":"颜色","values":["红","蓝"]},
  {"name":"尺寸","values":["S","M"]}
]`

// tpl 把 JSON 文本包成 JSONB 列的类型。
// 必须用这个而不是 []byte:入参会先经 json.Marshal 再 Unmarshal,
// 裸 []byte 会被二次编码成 JSON 字符串,反而解不出规格数组。
// 真实入参就是 datatypes.JSON(sys_product_spu.spec_template 的列类型)。
func tpl(s string) datatypes.JSON {
	return datatypes.JSON([]byte(s))
}

func skuWith(specValues map[string]interface{}) model.SysProductSku {
	m := datatypes.JSONMap{}
	for k, v := range specValues {
		m[k] = v
	}
	return model.SysProductSku{SpecValues: m}
}

// validSkuList 4 个 SKU,覆盖模板的全部笛卡尔积组合
func validSkuList() []model.SysProductSku {
	return []model.SysProductSku{
		skuWith(map[string]interface{}{"颜色": "红", "尺寸": "S"}),
		skuWith(map[string]interface{}{"颜色": "红", "尺寸": "M"}),
		skuWith(map[string]interface{}{"颜色": "蓝", "尺寸": "S"}),
		skuWith(map[string]interface{}{"颜色": "蓝", "尺寸": "M"}),
	}
}

func TestValidateSpecSkuRules(t *testing.T) {
	tests := []struct {
		name         string
		specTemplate interface{}
		skuList      []model.SysProductSku
		wantErr      error
	}{
		{
			name:         "合法模板与完整笛卡尔积通过",
			specTemplate: tpl(validSpecTemplate),
			skuList:      validSkuList(),
			wantErr:      nil,
		},
		{
			name:         "模板为空数组",
			specTemplate: tpl(`[]`),
			skuList:      validSkuList(),
			wantErr:      model.ErrSpuTemplate,
		},
		{
			name:         "模板为 nil",
			specTemplate: nil,
			skuList:      validSkuList(),
			wantErr:      model.ErrSpuTemplate,
		},
		{
			name:         "模板不是 JSON 数组",
			specTemplate: tpl(`{"name":"颜色","values":["红"]}`),
			skuList:      validSkuList(),
			wantErr:      model.ErrSpuTemplate,
		},
		{
			name: "规格缺 values",
			specTemplate: tpl(`[
			  {"name":"颜色","values":["红","蓝"]},
			  {"name":"尺寸","values":[]}
			]`),
			skuList: validSkuList(),
			wantErr: model.ErrSpuTemplate,
		},
		{
			name: "规格缺 name",
			specTemplate: tpl(`[
			  {"name":"","values":["红","蓝"]},
			  {"name":"尺寸","values":["S","M"]}
			]`),
			skuList: validSkuList(),
			wantErr: model.ErrSpuTemplate,
		},
		{
			name:         "SKU 数少于笛卡尔积",
			specTemplate: tpl(validSpecTemplate),
			skuList:      validSkuList()[:3],
			wantErr:      model.ErrSkuNum,
		},
		{
			name:         "SKU 数多于笛卡尔积",
			specTemplate: tpl(validSpecTemplate),
			skuList: append(validSkuList(), skuWith(map[string]interface{}{
				"颜色": "红", "尺寸": "L",
			})),
			wantErr: model.ErrSkuNum,
		},
		{
			name:         "规格键数量与模板不一致",
			specTemplate: tpl(validSpecTemplate),
			skuList: []model.SysProductSku{
				skuWith(map[string]interface{}{"颜色": "红"}),
				skuWith(map[string]interface{}{"颜色": "红"}),
				skuWith(map[string]interface{}{"颜色": "蓝"}),
				skuWith(map[string]interface{}{"颜色": "蓝"}),
			},
			wantErr: model.ErrSpecValues,
		},
		{
			name:         "规格名与模板不匹配",
			specTemplate: tpl(validSpecTemplate),
			skuList: []model.SysProductSku{
				skuWith(map[string]interface{}{"颜色": "红", "尺码": "S"}),
				skuWith(map[string]interface{}{"颜色": "红", "尺码": "M"}),
				skuWith(map[string]interface{}{"颜色": "蓝", "尺码": "S"}),
				skuWith(map[string]interface{}{"颜色": "蓝", "尺码": "M"}),
			},
			wantErr: model.ErrSpecValues,
		},
		{
			name:         "规格取值不在模板枚举内",
			specTemplate: tpl(validSpecTemplate),
			skuList: []model.SysProductSku{
				skuWith(map[string]interface{}{"颜色": "绿", "尺寸": "S"}),
				skuWith(map[string]interface{}{"颜色": "红", "尺寸": "M"}),
				skuWith(map[string]interface{}{"颜色": "蓝", "尺寸": "S"}),
				skuWith(map[string]interface{}{"颜色": "蓝", "尺寸": "M"}),
			},
			wantErr: model.ErrSpecValues,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateSpecSkuRules(tt.specTemplate, tt.skuList)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateSpecSkuRules() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// 模板以 Go 侧解出的 map 列表形态传入(局部更新的合并路径)也必须能解析
func TestValidateSpecSkuRules_AcceptsDecodedTemplate(t *testing.T) {
	decoded := []map[string]interface{}{
		{"name": "颜色", "values": []interface{}{"红", "蓝"}},
		{"name": "尺寸", "values": []interface{}{"S", "M"}},
	}
	if _, err := validateSpecSkuRules(decoded, validSkuList()); err != nil {
		t.Fatalf("已解码的模板应通过校验, got %v", err)
	}
}

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		name                    string
		page, pageSize, maxSize int
		wantPage, wantPageSize  int
	}{
		{"零值回落默认", 0, 0, 100, 1, 10},
		{"负数回落默认", -3, -5, 100, 1, 10},
		{"正常值原样", 2, 20, 100, 2, 20},
		{"超上限封顶", 1, 500, 100, 1, 100},
		{"恰好等于上限", 1, 100, 100, 1, 100},
		{"用户端上限 50", 1, 80, 50, 1, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := normalizePage(tt.page, tt.pageSize, tt.maxSize)
			if gotPage != tt.wantPage || gotPageSize != tt.wantPageSize {
				t.Fatalf("normalizePage(%d,%d,%d) = (%d,%d), want (%d,%d)",
					tt.page, tt.pageSize, tt.maxSize, gotPage, gotPageSize, tt.wantPage, tt.wantPageSize)
			}
		})
	}
}

// 业务错误归类:决定 gRPC 是走 error_msg(业务失败)还是 error(基础设施故障)。
// 归错会让调用方把"数据库抖动"当业务失败直接回 500 文案,或反之无限重试。
func TestIsProductBizError(t *testing.T) {
	bizCases := []error{
		model.ProductNotExist,
		model.ErrSkuNum,
		model.ErrSpecValues,
		model.ErrCategoryNotUsed,
		model.ErrPublishedCantChangeSpec,
		model.ErrSkuListEmpty,
		model.ErrInvalidStatusTransition,
		model.ErrNoActiveSku,
		model.ErrNoAvailableStock,
		model.ErrInvalidPrice,
		model.ErrSkuCodeNotOnly,
		model.ErrSpecValuesNotOnly,
		model.ErrSpuTemplate,
		model.ErrSpuDisabled,
	}
	for _, err := range bizCases {
		if !isProductBizError(err) {
			t.Errorf("业务错误未被归类: %v", err)
		}
	}

	// 类目域错误不应落进商品域集合(两域各自声明,避免误判)
	if isProductBizError(model.CategoryNotExist) {
		t.Error("类目域错误不应被商品域归类")
	}
	if isProductBizError(errors.New("connection refused")) {
		t.Error("基础设施故障不应被归类为业务错误")
	}
}
