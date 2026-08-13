import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

// 报销记录 API
export const reimbursementApi = {
  // 获取列表
  getList(params) {
    return api.get('/reimbursements', { params })
  },
  // 获取详情
  getDetail(id) {
    return api.get(`/reimbursements/${id}`)
  },
  // 创建
  create(data) {
    return api.post('/reimbursements', data)
  },
  // 更新
  update(id, data) {
    return api.put(`/reimbursements/${id}`, data)
  },
  // 删除
  delete(id) {
    return api.delete(`/reimbursements/${id}`)
  },
  // 更新状态
  updateStatus(id, status) {
    return api.patch(`/reimbursements/${id}/status`, { status })
  },
  // 更新人员状态
  updatePersonStatus(personId, data) {
    return api.patch(`/reimbursements/persons/${personId}/status`, data)
  },
  // 获取人员姓名列表
  getPersonNames() {
    return api.get('/persons/names')
  },
  // 获取报销缘由选项
  getReasonOptions() {
    return api.get('/reasons/options')
  },
  // 获取统计数据
  getStatistics() {
    return api.get('/statistics')
  }
}

export default api
