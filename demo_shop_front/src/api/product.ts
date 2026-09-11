import {CreateProductReq, SpuQueryReq, UpdateProductFullReq, UpdateProductReq} from "@/types/product";
import service from "@/utils/request";


export const CreateProductApi  = (data: CreateProductReq = {}) => {
    return service({
        url: '/platform/products',
        method: 'post',
        data: data
    })
}


export const GetProductListApi = (params : SpuQueryReq = {}) => {
    return service({
        url : '/platform/products',
        method : 'get',
        params
    })
}

export const GetProductApi = (id:number) => {
    return service({
        url : `/platform/products/${id}`,
        method : 'get',
    })
}

export const UpdateProductApi = (id:number, data: UpdateProductReq) => {
    return service({
        url: `/platform/products/${id}`,
        method: 'put',
        data: data
    })
}

export const UpdateProductFullApi = (id:number, data: UpdateProductFullReq) => {
    return service({
        url: `/platform/products/${id}/full`,
        method: 'put',
        data: data
    })
}

export const DeleteProductApi = (id:number) => {
    return service({
        url: `/platform/products/${id}`,
        method: 'delete',
    })
}

export const PublishProductApi = (id:number) => {
    return service({
        url: `/platform/products/${id}/publish`,
        method: 'post',
    })
}

export const WithdrawProductApi = (id:number) => {
    return service({
        url: `/platform/products/${id}/withdraw`,
        method: 'post',
    })
}

export const UserGetProductListApi = (params : SpuQueryReq = {}) => {
    return service({
        url : '/users/platform/products',
        method : 'get',
        params
    })
}

export const UserGetProductApi = (id:number) => {
    return service({
        url : `/users/platform/products/${id}`,
        method : 'get',
    })
}


