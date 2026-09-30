import service from '@/utils/request'

// 菜单树按当前登录用户返回:用户身份取自 JWT,不接受请求体指定
export const asyncMenu = () => {
    return service({
        url: '/admin/menu/tree',
        method: 'post'
    })
}
