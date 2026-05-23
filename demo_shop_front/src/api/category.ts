import service from '@/utils/request'

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

export interface GetCategoryListParams{
    page? : number,
    pageSize?:number,
}



export const GetCategoryTreeApi = (params: CategoryTreeParams = {}) => {
    return service({
        url: '/api/v1/platform/category/tree',
        method: 'get',
        params
    })
}

export const CreateCategoryApi = (data:CreateCategoryData) => {
    return service({
        url: '/api/v1/platform/category',
        method: 'post',
        data: data
    })
}


export const UpdateCategoryApi = (id,data:UpdateCategoryData) => {
    return service({
        url: `/api/v1/platform/category/${id}`,
        method: 'put',
        data: data
    })
}


export const DeleteCategoryApi = (id) => {
    return service({
        url: `/api/v1/platform/category/${id}`,
        method: 'delete',
    })
}


export const GetCategoryApi = (id) => {
    return service({
        url: `/api/v1/platform/category/${id}`,
        method: 'get',
    })
}


export const GetCategoryChildrenListApi = (id,params: CategoryChildrenParams = {}) => {
    return service({
        url: `/api/v1/platform/category/children/${id}`,
        method: 'get',
        params
    })
}



