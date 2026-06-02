import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
    Category,
    CreateCategoryData,
    UpdateCategoryData,
    CategoryTreeParams,
    CategoryChildrenListParams
} from '@/types/category'
import {
    GetCategoryTreeApi,
    CreateCategoryApi,
    UpdateCategoryApi,
    DeleteCategoryApi,
    GetCategoryApi,
    GetCategoryChildrenListApi
} from '@/api/category'

export const useCategoryStore = defineStore('category', () => {
    // 存储类目树结构
    const categoryTree = ref<Category[]>([])
    // 存储类目信息
    const currentCategory = ref<Category | null>(null)
    // 存储类目的子类目列表
    const categoryChildren = ref<Category[]>([])

    // 异步获取类目树结构
    async function GetCategoryTree(params? : CategoryTreeParams):Promise<Category[]> {
        try{
            // 调用api获取类目树信息
            const res = await GetCategoryTreeApi(params)
            // 存储类目树
            categoryTree.value = res.data.data
            // 返回类目树结构信息
            return res.data.data
        } catch (error) {
            console.log("获取类目树失败")
            return []
        }
    }

    // 异步获取单个类目信息
    async function GetCategory(id : number):Promise<Category | null>{
        try{
            // 调用api获取单个类目信息
            const res = await GetCategoryApi(id)
            // 存储单个类目
            currentCategory.value = res.data.data
            // 返回单个类目信息
            return res.data.data
        } catch (error) {
            console.log("获取类目详情失败")
            return null
        }
    }

    // 异步获取类目的子类目列表
    async function GetCategoryChildren(id:number, params?:CategoryChildrenListParams):Promise<Category[]>{
        try{
            // 调用api获取类目的子类目列表信息
            const res = await GetCategoryChildrenListApi(id, params)
            // 存储类目的子类目列表信息
            categoryChildren.value = res.data.data
            // 返回类目的子类目列表信息
            return res.data.data
        }catch (error){
            console.log("获取类目子列表失败")
            return []
        }
    }

    // 创建新类目
    async function CreateCategory(data:CreateCategoryData):Promise<Category>{
        try{
            const res = await CreateCategoryApi(data)
            console.log("创建类目成功")
            await GetCategoryTree()
            return res.data.data
        }catch (error){
            console.log("创建类目失败")
            return null
        }
    }
    // 更新类目信息
    async function UpdateCategory(id:number, data:UpdateCategoryData):Promise<Category>{
        try{
            const res = await UpdateCategoryApi(id, data)
            await Promise.all([
                GetCategoryTree(),
                currentCategory.value?.category_id === id ? GetCategory(id) : Promise.resolve()
            ])
            return res.data.data
        }catch (error){
            console.log("更新类目失败")
            return null
        }
    }
    // 删除类目
    async function DeleteCategory(id:number):Promise<boolean>{
        try{
            await DeleteCategoryApi(id)

            await GetCategoryTree()
            // 如果删除的是当前选中的类目，清空当前选中状态
            if (currentCategory.value?.category_id === id) {
                currentCategory.value = null
            }
            return true
        } catch (error) {
            console.log("删除类目失败")
            return false
        }
    }

    // 清除类目信息
    function resetCurrentCategory() {
        currentCategory.value = null
    }

    // 清除类目的子类目列表信息
    function resetCategoryChildren(){
        categoryChildren.value = []
    }

    return{
        // 状态
        categoryTree,
        currentCategory,
        categoryChildren,

        // 方法
        GetCategoryTree,
        GetCategory,
        GetCategoryChildren,
        CreateCategory,
        UpdateCategory,
        DeleteCategory,
        resetCategoryChildren,
        resetCurrentCategory
    }


})