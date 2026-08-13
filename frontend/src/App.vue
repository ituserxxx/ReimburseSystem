<template>
  <div class="app-container">
    <!-- 顶部标题栏 -->
    <el-header class="app-header">
      <div class="header-left">
        <el-icon :size="28" color="#409EFF"><Wallet /></el-icon>
        <span class="app-title">报销记录管理系统</span>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleAdd">
          新增报销记录
        </el-button>
      </div>
    </el-header>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon" style="background: #ecf5ff;">
            <el-icon :size="24" color="#409EFF"><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">总记录数</div>
            <div class="stat-value">{{ stats.total_records || 0 }}</div>
          </div>
        </div>
      </el-card>
      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon" style="background: #fdf6ec;">
            <el-icon :size="24" color="#E6A23C"><Clock /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">待报销笔数</div>
            <div class="stat-value">{{ stats.pending_count || 0 }}</div>
          </div>
        </div>
      </el-card>
      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon" style="background: #fef0f0;">
            <el-icon :size="24" color="#F56C6C"><Money /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">待报销金额</div>
            <div class="stat-value">¥{{ formatAmount(stats.pending_amount) }}</div>
          </div>
        </div>
      </el-card>
      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon" style="background: #f0f9eb;">
            <el-icon :size="24" color="#67C23A"><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">已到账金额</div>
            <div class="stat-value">¥{{ formatAmount(stats.received_amount) }}</div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 记录列表 -->
    <RecordList
      ref="listRef"
      @edit="handleEdit"
      @detail="handleDetail"
      @status-change="handleStatusChange"
      @person-status-change="handlePersonStatusChange"
    />

    <!-- 新增/编辑表单 -->
    <RecordForm
      v-model:visible="formVisible"
      :edit-data="currentEditData"
      :person-names="personNames"
      :reason-options="reasonOptions"
      @saved="handleSaved"
    />

    <!-- 详情弹窗 -->
    <RecordDetail
      v-model:visible="detailVisible"
      :data="currentDetailData"
      @edit="handleEditFromDetail"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, Wallet, Document, Clock, Money, CircleCheck } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RecordList from './components/RecordList.vue'
import RecordForm from './components/RecordForm.vue'
import RecordDetail from './components/RecordDetail.vue'
import { reimbursementApi } from './api'

const listRef = ref()
const formVisible = ref(false)
const detailVisible = ref(false)
const currentEditData = ref(null)
const currentDetailData = ref(null)
const stats = ref({})
const personNames = ref([])
const reasonOptions = ref([])

// 格式化金额
const formatAmount = (val) => {
  return (val || 0).toFixed(2)
}

// 加载统计数据
const loadStats = async () => {
  try {
    const res = await reimbursementApi.getStatistics()
    stats.value = res.data
  } catch (e) {
    console.error('加载统计失败', e)
  }
}

// 加载人员姓名和缘由选项
const loadOptions = async () => {
  try {
    const [nameRes, reasonRes] = await Promise.all([
      reimbursementApi.getPersonNames(),
      reimbursementApi.getReasonOptions()
    ])
    personNames.value = nameRes.data || []
    reasonOptions.value = reasonRes.data || []
  } catch (e) {
    console.error('加载选项失败', e)
  }
}

// 新增
const handleAdd = () => {
  currentEditData.value = null
  formVisible.value = true
}

// 编辑
const handleEdit = (row) => {
  currentEditData.value = row
  formVisible.value = true
}

// 详情
const handleDetail = (row) => {
  currentDetailData.value = row
  detailVisible.value = true
}

// 从详情页编辑
const handleEditFromDetail = (row) => {
  detailVisible.value = false
  currentEditData.value = row
  formVisible.value = true
}

// 保存成功后刷新
const handleSaved = () => {
  formVisible.value = false
  listRef.value?.loadData()
  loadStats()
  loadOptions()
}

// 整体状态变更
const handleStatusChange = async ({ id, status }) => {
  try {
    await ElMessageBox.confirm(
      `确认将此记录标记为「${status === 'received' ? '已到账' : '待报销'}」？`,
      '提示',
      { type: 'warning' }
    )
    await reimbursementApi.updateStatus(id, status)
    ElMessage.success('状态更新成功')
    listRef.value?.loadData()
    loadStats()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('状态更新失败')
    }
  }
}

// 单人状态变更
const handlePersonStatusChange = async ({ personId, data }) => {
  try {
    await reimbursementApi.updatePersonStatus(personId, data)
    ElMessage.success('人员状态更新成功')
    listRef.value?.loadData()
    loadStats()
  } catch (e) {
    ElMessage.error('状态更新失败')
  }
}

onMounted(() => {
  loadStats()
  loadOptions()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
  background-color: #f5f7fa;
}

.app-container {
  min-height: 100vh;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  padding: 0 24px;
  height: 64px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.header-right {
  display: flex;
  gap: 12px;
}

.stats-row {
  display: flex;
  gap: 16px;
  padding: 20px 24px;
}

.stat-card {
  flex: 1;
  min-width: 200px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

/* 覆盖 el-card body padding */
.stat-card .el-card__body {
  padding: 16px 20px;
}

.main-content {
  padding: 0 24px 24px;
}
</style>
