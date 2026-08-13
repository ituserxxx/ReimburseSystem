<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑报销记录' : '新增报销记录'"
    width="780px"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      label-position="right"
    >
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="报销人" prop="submitter">
            <el-autocomplete
              v-model="form.submitter"
              :fetch-suggestions="queryPersonNames"
              placeholder="请输入报销人姓名"
              style="width: 100%"
              clearable
              @input="onSubmitterInput"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="地点" prop="location">
            <el-input
              v-model="form.location"
              placeholder="请输入报销地点"
              clearable
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="提交日期" prop="submit_date">
            <el-date-picker
              v-model="form.submit_date"
              type="date"
              placeholder="选择日期"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="总金额" prop="total_amount">
            <el-input-number
              v-model="form.total_amount"
              :precision="2"
              :min="0"
              :step="100"
              controls-position="right"
              style="width: 100%"
              @change="onTotalChange"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="是否多人">
        <el-switch
          v-model="form.is_multi_person"
          active-text="多人报销"
          inactive-text="单人报销"
          @change="handleMultiChange"
        />
      </el-form-item>

      <el-form-item label="报销缘由" prop="reasons">
        <el-select
          v-model="form.reasons"
          multiple
          placeholder="选择报销缘由（可多选）"
          style="width: 100%"
          allow-create
          filterable
          default-first-option
        >
          <el-option
            v-for="opt in reasonOptions"
            :key="opt"
            :label="opt"
            :value="opt"
          />
        </el-select>
      </el-form-item>

      <!-- 人员分配：单人 / 多人 统一使用表格样式；单人锁定为 1 人 -->
      <el-form-item label="人员分配">
        <div class="persons-section">
          <el-table :data="form.persons" border size="small" style="width: 100%">
            <el-table-column label="姓名" min-width="120">
              <template #default="{ row }">
                <el-autocomplete
                  v-model="row.person_name"
                  :fetch-suggestions="queryPersonNames"
                  placeholder="姓名"
                  size="small"
                  style="width: 100%"
                />
              </template>
            </el-table-column>
            <el-table-column label="金额" width="130">
              <template #default="{ row }">
                <el-input-number
                  v-model="row.amount"
                  :precision="2"
                  :min="0"
                  :step="100"
                  :controls="false"
                  size="small"
                  style="width: 100%"
                  @change="onPersonAmountChange"
                />
              </template>
            </el-table-column>
            <el-table-column label="审批单号" min-width="140">
              <template #default="{ row }">
                <el-input
                  v-model="row.approval_number"
                  placeholder="审批单号"
                  size="small"
                  clearable
                />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="70" align="center">
              <template #default="{ $index }">
                <el-button
                  type="danger"
                  size="small"
                  link
                  :icon="Delete"
                  @click="removePerson($index)"
                  :disabled="!form.is_multi_person || form.persons.length <= 1"
                />
              </template>
            </el-table-column>
          </el-table>

          <div class="persons-footer">
            <el-button
              v-if="form.is_multi_person"
              size="small"
              type="primary"
              plain
              :icon="Plus"
              @click="addPerson"
            >
              添加人员
            </el-button>
            <div class="persons-summary">
              <span>已分配: ¥{{ formatAmount(personsTotalAllocated) }}</span>
              <span>总金额: ¥{{ formatAmount(form.total_amount) }}</span>
              <el-tag
                v-if="form.is_multi_person && !isAllocatedMatch"
                type="danger"
                size="small"
              >
                金额不匹配 (差 ¥{{ formatAmount(form.total_amount - personsTotalAllocated) }})
              </el-tag>
              <el-tag
                v-else-if="form.is_multi_person"
                type="success"
                size="small"
              >
                金额匹配
              </el-tag>
              <el-tag v-else type="info" size="small">
                单人报销（固定 1 人）
              </el-tag>
            </div>
          </div>
        </div>
      </el-form-item>

      <el-form-item label="备注">
        <el-input
          v-model="form.remark"
          type="textarea"
          :rows="2"
          placeholder="备注信息（选填）"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { reimbursementApi } from '../api'

const props = defineProps({
  visible: Boolean,
  editData: Object,
  personNames: Array,
  reasonOptions: Array
})

const emit = defineEmits(['update:visible', 'saved'])

const formRef = ref()
const saving = ref(false)

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const isEdit = computed(() => !!props.editData)

const defaultForm = () => ({
  submitter: '',
  location: '',
  is_multi_person: false,
  total_amount: 0,
  submit_date: new Date().toISOString().split('T')[0],
  remark: '',
  persons: [emptyPerson()],
  reasons: []
})

function emptyPerson() {
  return {
    person_name: '',
    amount: 0,
    approval_number: '',
    status: 'pending'
  }
}

const form = reactive(defaultForm())

const rules = {
  submitter: [{ required: true, message: '请输入报销人', trigger: 'blur' }],
  location: [{ required: true, message: '请输入地点', trigger: 'blur' }],
  total_amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
  submit_date: [{ required: true, message: '请选择日期', trigger: 'change' }]
}

// 人员总分配金额
const personsTotalAllocated = computed(() => {
  return (form.persons || []).reduce((sum, p) => sum + (Number(p.amount) || 0), 0)
})

const isAllocatedMatch = computed(() => {
  if (!form.is_multi_person) return true
  return Math.abs(personsTotalAllocated.value - form.total_amount) < 0.01
})

const formatAmount = (val) => Number(val || 0).toFixed(2)

// 自动完成搜索
const queryPersonNames = (queryString, cb) => {
  const names = props.personNames || []
  const results = queryString
    ? names.filter(n => n.toLowerCase().includes(queryString.toLowerCase()))
    : names
  cb(results.map(n => ({ value: n })))
}

// 报销人输入时，单人模式仅在人员姓名为空时同步（保留用户手填）
const onSubmitterInput = () => {
  if (!form.is_multi_person && form.persons[0] && !form.persons[0].person_name) {
    form.persons[0].person_name = form.submitter
  }
}

// 总金额变化
const onTotalChange = () => {
  if (!form.is_multi_person && form.persons[0]) {
    form.persons[0].amount = form.total_amount
  } else if (form.is_multi_person && form.persons.length > 0) {
    autoDistribute()
  }
}

// 人员金额变化：单人模式下反向同步回总金额
const onPersonAmountChange = () => {
  if (!form.is_multi_person && form.persons[0]) {
    form.total_amount = form.persons[0].amount
  }
}

// 多人切换
const handleMultiChange = (val) => {
  if (val) {
    // 切到多人：确保至少有一人
    if (form.persons.length === 0) {
      form.persons = [emptyPerson()]
    }
    // 同步第一人信息（来自单人时的报销人）
    form.persons[0].person_name = form.persons[0].person_name || form.submitter
    form.persons[0].amount = form.persons[0].amount || form.total_amount
  } else {
    // 切到单人：固定为报销人一人，金额=总额
    form.persons = [{
      person_name: form.submitter,
      amount: form.total_amount,
      approval_number: form.persons[0]?.approval_number || '',
      status: 'pending'
    }]
  }
}

// 自动平均分配金额（多人）
const autoDistribute = () => {
  const count = form.persons.length
  if (count === 0) return
  const base = Math.floor((form.total_amount / count) * 100) / 100
  const remainder = Math.round((form.total_amount - base * count) * 100) / 100
  form.persons.forEach((p, i) => {
    p.amount = i === 0 ? base + remainder : base
  })
}

// 添加人员
const addPerson = () => {
  form.persons.push(emptyPerson())
}

// 删除人员
const removePerson = (index) => {
  if (form.persons.length <= 1) return
  form.persons.splice(index, 1)
}

// 重置表单
const resetForm = () => {
  Object.assign(form, defaultForm())
  formRef.value?.resetFields()
}

// 填充编辑数据
const fillEditData = (data) => {
  if (!data) {
    resetForm()
    return
  }
  form.submitter = data.submitter || ''
  form.location = data.location || ''
  form.is_multi_person = data.is_multi_person || false
  form.total_amount = data.total_amount || 0
  form.submit_date = data.submit_date || ''
  form.remark = data.remark || ''
  form.reasons = (data.reasons || []).map(r => r.reason)
  form.persons = (data.persons && data.persons.length > 0)
    ? data.persons.map(p => ({
        person_name: p.person_name,
        amount: p.amount,
        approval_number: p.approval_number || '',
        status: p.status || 'pending'
      }))
    : [emptyPerson()]
}

// 监听弹窗打开
watch(() => props.visible, (val) => {
  if (val) {
    if (props.editData) {
      fillEditData(props.editData)
    } else {
      resetForm()
    }
  }
})

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()

    // 单人模式固定一人（保留用户填写的姓名/金额，仅在为空时回退默认值）
    if (!form.is_multi_person) {
      if (!form.persons[0]) form.persons = [emptyPerson()]
      const p = form.persons[0]
      if (!p.person_name?.trim()) {
        ElMessage.warning('请填写人员姓名')
        return
      }
      form.persons = [{
        person_name: p.person_name.trim(),
        amount: p.amount !== undefined && p.amount !== null ? p.amount : form.total_amount,
        approval_number: p.approval_number || '',
        status: 'pending'
      }]
    } else {
      // 多人模式校验
      if (form.persons.length === 0) {
        ElMessage.warning('多人报销请至少添加一名人员')
        return
      }
      for (const p of form.persons) {
        if (!p.person_name?.trim()) {
          ElMessage.warning('请填写所有人员姓名')
          return
        }
        if (!p.amount || p.amount <= 0) {
          ElMessage.warning('请填写所有人员金额')
          return
        }
      }
      if (!isAllocatedMatch.value) {
        ElMessage.warning('人员分配金额与总金额不匹配')
        return
      }
    }

    saving.value = true

    const payload = {
      submitter: form.submitter,
      location: form.location,
      is_multi_person: form.is_multi_person,
      total_amount: form.total_amount,
      submit_date: form.submit_date,
      remark: form.remark,
      reasons: form.reasons,
      persons: form.persons.map(p => ({
        person_name: p.person_name,
        amount: p.amount,
        approval_number: p.approval_number || '',
        status: p.status || 'pending'
      }))
    }

    if (isEdit.value) {
      await reimbursementApi.update(props.editData.id, payload)
      ElMessage.success('更新成功')
    } else {
      await reimbursementApi.create(payload)
      ElMessage.success('创建成功')
    }
    emit('saved')
  } catch (e) {
    if (e?.message) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
    }
  } finally {
    saving.value = false
  }
}

const handleClosed = () => {
  resetForm()
}
</script>

<style scoped>
.persons-section {
  width: 100%;
}

.persons-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  flex-wrap: wrap;
  gap: 8px;
}

.persons-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: #606266;
}
</style>
