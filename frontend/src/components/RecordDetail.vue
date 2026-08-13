<template>
  <el-dialog
    v-model="dialogVisible"
    title="报销记录详情"
    width="720px"
    @closed="handleClosed"
  >
    <div v-if="data" class="detail-container">
      <!-- 基本信息 -->
      <el-descriptions :column="2" border>
        <el-descriptions-item label="报销人">
          <span class="detail-value">{{ data.submitter }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="地点">
          {{ data.location }}
        </el-descriptions-item>
        <el-descriptions-item label="提交日期">
          {{ data.submit_date }}
        </el-descriptions-item>
        <el-descriptions-item label="总金额">
          <span class="amount-highlight">¥{{ formatAmount(data.total_amount) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="报销类型">
          <el-tag :type="data.is_multi_person ? 'warning' : 'info'" size="small">
            {{ data.is_multi_person ? '多人报销' : '单人报销' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="整体状态">
          <el-tag
            :type="data.status === 'received' ? 'success' : 'warning'"
            effect="light"
          >
            {{ data.status === 'received' ? '已到账' : '待报销' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(data.created_at) }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 报销缘由 -->
      <div class="section-title">报销缘由</div>
      <div class="reasons-list">
        <el-tag
          v-for="reason in (data.reasons || [])"
          :key="reason.id"
          :type="getReasonTagType(reason.reason)"
          size="default"
          style="margin-right: 8px; margin-bottom: 4px"
        >
          {{ reason.reason }}
        </el-tag>
        <span v-if="!data.reasons || data.reasons.length === 0" class="empty-text">
          无
        </span>
      </div>

      <!-- 人员明细 -->
      <div class="section-title">
        人员明细
        <span v-if="data.is_multi_person" class="section-hint">
          （共 {{ (data.persons || []).length }} 人）
        </span>
      </div>
      <el-table
        :data="data.persons || []"
        border
        size="small"
        style="width: 100%"
        :header-cell-style="{ background: '#f5f7fa', fontWeight: 600 }"
      >
        <el-table-column label="姓名" prop="person_name" min-width="100" />
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">
            ¥{{ formatAmount(row.amount) }}
          </template>
        </el-table-column>
        <el-table-column label="审批单号" min-width="150">
          <template #default="{ row }">
            <span v-if="row.approval_number">{{ row.approval_number }}</span>
            <span v-else class="empty-text">未填写</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'received' ? 'success' : 'warning'"
              size="small"
            >
              {{ row.status === 'received' ? '已到账' : '待报销' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <!-- 备注 -->
      <div v-if="data.remark" class="section-title">备注</div>
      <div v-if="data.remark" class="remark-box">
        {{ data.remark }}
      </div>
    </div>

    <template #footer>
      <el-button @click="dialogVisible = false">关闭</el-button>
      <el-button type="primary" @click="$emit('edit', data)">编辑</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  visible: Boolean,
  data: Object
})

const emit = defineEmits(['update:visible', 'edit'])

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const formatAmount = (val) => Number(val || 0).toFixed(2)

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
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

const getReasonTagType = (reason) => reasonTagMap[reason] || 'info'

const handleClosed = () => {}
</script>

<style scoped>
.detail-container {
  max-height: 600px;
  overflow-y: auto;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin: 20px 0 10px;
}

.section-hint {
  font-size: 13px;
  color: #909399;
  font-weight: normal;
}

.detail-value {
  font-weight: 500;
}

.amount-highlight {
  color: #f56c6c;
  font-weight: 700;
  font-size: 16px;
}

.reasons-list {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}

.remark-box {
  background: #f5f7fa;
  padding: 12px 16px;
  border-radius: 4px;
  color: #606266;
  line-height: 1.6;
}

.empty-text {
  color: #c0c4cc;
  font-size: 13px;
}
</style>
