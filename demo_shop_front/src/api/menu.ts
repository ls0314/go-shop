import service from '@/utils/request'

export const asyncMenu = (data) => {
    return service({
        url: '/admin/menu/tree',
        method: 'post',
        data: data
    })
}