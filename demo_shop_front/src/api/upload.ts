import service from '@/utils/request'
import SparkMD5 from 'spark-md5'

export interface UploadResp {
    is_completed: boolean
    file_url: string
}

/** 上传单个文件，返回文件 URL */
export function uploadFile(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onload = (e) => {
            const spark = new SparkMD5.ArrayBuffer()
            spark.append(e.target?.result as ArrayBuffer)
            const md5 = spark.end()

            const formData = new FormData()
            formData.append('file_md5', md5)
            formData.append('file_name', file.name)
            formData.append('chunk_number', '1')
            formData.append('total_chunks', '1')
            formData.append('file_size', String(file.size))
            formData.append('file', file)

            service({
                url: '/api/v1/upload/chunk',
                method: 'post',
                data: formData,
                headers: { 'Content-Type': 'multipart/form-data' },
            })
                .then((res) => {
                    const data = res.data.data as UploadResp
                    if (data?.file_url) {
                        resolve(data.file_url)
                    } else {
                        reject(new Error('上传未完成'))
                    }
                })
                .catch(reject)
        }
        reader.readAsArrayBuffer(file)
    })
}
