<template>
  <div class="main-content">
    <el-card shadow="never">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-input
          v-model="filters.submitter"
          placeholder="搜索报销人"
          clearable
          style="width: 180px"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="filters.status"
          placeholder="报销状态"
          clearable
          style="width: 140px"
          @change="handleSearch"
        >
          <el-option label="待报销" value="pending" />
          <el-option label="已到账" value="received" />
        </el-select>
        <el-select
          v-model="filters.is_multi_person"
          placeholder="报销类型"
          clearable
          style="width: 140px"
          @change="handleSearch"
        >
          <el-option label="单人报销" value="false" />
          <el-option label="多人报销" value="true" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 260px"
          @change="handleDateChange"
        />
        <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
        <el-button :icon="Refresh" @click="handleReset">重置</el-button>
      </div>

      <!-- 表格 -->
      <el-table
        :data="tableData"
        v-loading="loading"
        border
        stripe
        style="width: 100%; margin-top: 16px"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: 600 }"
      >
        <el-table-column label="序号" type="index" width="60" align="center" />

        <el-table-column label="报销人" prop="submitter" width="100" />

        <el-table-column label="地点" prop="location" width="120" show-overflow-tooltip />

        <el-table-column label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_multi_person ? 'warning' : 'info'" size="small">
              {{ row.is_multi_person ? '多人' : '单人' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="金额" width="110" align="right">
          <template #default="{ row }">
            <span class="amount-text">¥{{ formatAmount(row.total_amount) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="报销缘由" min-width="160">
          <template #default="{ row }">
            <el-tag
              v-for="reason in (row.reasons || [])"
              :key="reason.id"
              :type="getReasonTagType(reason.reason)"
              size="small"
              style="margin-right: 4px; margin-bottom: 2px"
            >
              {{ reason.reason }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="人员明细" min-width="200">
          <template #default="{ row }">
            <div v-if="row.is_multi_person && row.persons && row.persons.length > 0">
              <div
                v-for="person in row.persons"
                :key="person.id"
                class="person-row"
              >
                <span class="person-name">{{ person.person_name }}</span>
                <span class="person-amount">¥{{ formatAmount(person.amount) }}</span>
                <el-tag
                  :type="person.status === 'received' ? 'success' : 'warning'"
                  size="small"
                >
                  {{ person.status === 'received' ? '已到账' : '待报销' }}
                </el-tag>
              </div>
            </div>
            <div v-else class="single-person">
              <span class="person-name">{{ row.submitter }}</span>
              <el-tag
                :type="row.status === 'received' ? 'success' : 'warning'"
                size="small"
              >
                {{ row.status === 'received' ? '已到账' : '待报销' }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="提交日期" prop="submit_date" width="110" align="center" />

        <el-table-column label="提交时间" width="160" align="center">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'received' ? 'success' : 'warning'"
              effect="light"
            >
              {{ row.status === 'received' ? '已到账' : '待报销' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, row)">
              <el-button size="small" type="primary" plain>
                操作<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="detail" :icon="View">详情</el-dropdown-item>
                  <el-dropdown-item command="edit" :icon="Edit">编辑</el-dropdown-item>
                  <el-dropdown-item
                    v-if="row.status === 'pending'"
                    command="received"
                    :icon="Check"
                  >
                    标记到账
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-else
                    command="pending"
                    :icon="RefreshLeft"
                  >
                    撤销到账
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" :icon="Delete" divided>
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import {
  Search,
  Refresh,
  Delete,
  ArrowDown,
  View,
  Edit,
  Check,
  RefreshLeft
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { reimbursementApi } from '../api'

const emit = defineEmits(['edit', 'detail', 'status-change', 'person-status-change'])

const loading = ref(false)
const tableData = ref([])
const dateRange = ref([])

const filters = reactive({
  submitter: '',
  status: '',
  is_multi_person: '',
  start_date: '',
  end_date: ''
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

const formatAmount = (val) => {
  return Number(val || 0).toFixed(2)
}

// 格式化提交时间（created_at: ISO 字符串 -> YYYY-MM-DD HH:mm:ss）
const formatTime = (val) => {
  if (!val) return '-'
  const d = new Date(val)
  if (isNaN(d.getTime())) return val
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 操作下拉按钮命令分发
const handleCommand = (command, row) => {
  switch (command) {
    case 'detail':
      emit('detail', row)
      break
    case 'edit':
      emit('edit', row)
      break
    case 'received':
      emit('status-change', { id: row.id, status: 'received' })
      break
    case 'pending':
      emit('status-change', { id: row.id, status: 'pending' })
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

const reasonTagMap = {
  '交通': '',
  'ETC': 'success',
  '油费': 'warning',
  '误餐': 'danger',
  '住宿': 'info',
  '办公用品': '',
  '其他': 'info'
}

const getReasonTagType = (reason) => {
  return reasonTagMap[reason] || 'info'
}

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      ...filters
    }
    // 清理空值
    Object.keys(params).forEach(k => {
      if (params[k] === '' || params[k] === null) delete params[k]
    })

    const res = await reimbursementApi.getList(params)
    tableData.value = res.data.list || []
    pagination.total = res.data.total || 0
  } catch (e) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  filters.submitter = ''
  filters.status = ''
  filters.is_multi_person = ''
  filters.start_date = ''
  filters.end_date = ''
  dateRange.value = []
  pagination.page = 1
  loadData()
}

const handleDateChange = (val) => {
  if (val && val.length === 2) {
    filters.start_date = val[0]
    filters.end_date = val[1]
  } else {
    filters.start_date = ''
    filters.end_date = ''
  }
  handleSearch()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认删除报销人「${row.submitter}」于 ${row.submit_date} 提交的报销记录？`,
      '删除确认',
      { type: 'warning' }
    )
    await reimbursementApi.delete(row.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

defineExpose({
  loadData
})

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
}

.amount-text {
  font-weight: 600;
  color: #303133;
}

.person-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 2px 0;
}

.single-person {
  display: flex;
  align-items: center;
  gap: 8px;
}

.person-name {
  color: #303133;
  font-size: 13px;
  min-width: 40px;
}

.person-amount {
  color: #606266;
  font-size: 13px;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
