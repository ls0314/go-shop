import service from '@/utils/request'
import type { CategoryTreeParams, CreateCategoryData, UpdateCategoryData, CategoryChildrenParams } from '@/types/category'

export const GetCategoryTreeApi = (params: CategoryTreeParams = {}) => {
    return service({
        url: '/platform/category/tree',
        method: 'get',
        params
    })
}

export const CreateCategoryApi = (data: CreateCategoryData) => {
    return service({
        url: '/platform/category',
        method: 'post',
        data: data
    })
}

export const UpdateCategoryApi = (id, data: UpdateCategoryData) => {
    return service({
        url: `/platform/category/${id}`,
        method: 'put',
        data: data
    })
}

export const DeleteCategoryApi = (id) => {
    return service({
        url: `/platform/category/${id}`,
        method: 'delete',
    })
}

export const GetCategoryApi = (id) => {
    return service({
        url: `/platform/category/${id}`,
        method: 'get',
    })
}

export const GetCategoryChildrenListApi = (id, params: CategoryChildrenParams = {}) => {
    return service({
        url: `/platform/category/children/${id}`,
        method: 'get',
        params
    })
}
