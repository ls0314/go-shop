// 类目相关类型定义

export interface Category {
    category_id: number
    parent_id: number
    category_name: string
    category_level: number
    category_path: string
    sort_order: number
    icon_url?: string
    is_leaf: boolean
    is_visible: boolean
    status: 'active' | 'disabled'
    created_at: string
    updated_at: string
    children?: Category[]
}

export interface CreateCategoryData {
    parent_id: number
    category_name: string
    sort_order?: number
    icon_url?: string
    is_visible?: boolean
}

export interface UpdateCategoryData {
    category_name?: string
    sort_order?: number
    icon_url?: string
    is_visible?: boolean
    status?: 'active' | 'disabled'
}

export interface CategoryTreeParams {
    level?: number
    include_disabled?: boolean
}

export interface CategoryChildrenParams {
    include_disabled?: boolean
}

export type CategoryChildrenListParams = CategoryChildrenParams

export interface GetCategoryListParams {
    page?: number
    pageSize?: number
}
