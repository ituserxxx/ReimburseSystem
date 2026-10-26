// 截图脚本：拦截 /api 请求返回 mock 数据，截取前台各页面
const { chromium } = require('playwright')

const OUT = __dirname + '/screenshots'

const mockList = {
  code: 0,
  data: {
    total: 4,
    list: [
      {
        id: 4, submitter: '张三', location: '深圳北站', is_multi_person: true,
        total_amount: 1200.00, submit_date: '2026-10-20', status: 'pending',
        remark: '项目组出差返程', created_at: '2026-10-20T14:32:11+08:00',
        persons: [
          { id: 7, person_name: '张三', amount: 400, approval_number: 'SP20261020001', status: 'pending' },
          { id: 8, person_name: '李四', amount: 400, approval_number: 'SP20261020002', status: 'pending' },
          { id: 9, person_name: '王五', amount: 400, approval_number: 'SP20261020003', status: 'pending' }
        ],
        reasons: [{ id: 10, reason: '交通' }, { id: 11, reason: 'ETC' }]
      },
      {
        id: 3, submitter: '李四', location: '广州南站', is_multi_person: false,
        total_amount: 356.50, submit_date: '2026-10-18', status: 'received',
        remark: '', created_at: '2026-10-18T09:05:40+08:00',
        persons: [
          { id: 5, person_name: '李四', amount: 356.5, approval_number: 'SP20261018001', status: 'received' }
        ],
        reasons: [{ id: 7, reason: '油费' }]
      },
      {
        id: 2, submitter: '张三', location: '东莞工厂', is_multi_person: true,
        total_amount: 680.00, submit_date: '2026-10-15', status: 'received',
        remark: '客户现场调试', created_at: '2026-10-15T16:20:03+08:00',
        persons: [
          { id: 3, person_name: '张三', amount: 340, approval_number: 'SP20261015001', status: 'received' },
          { id: 4, person_name: '赵六', amount: 340, approval_number: 'SP20261015002', status: 'received' }
        ],
        reasons: [{ id: 4, reason: '交通' }, { id: 5, reason: '误餐' }]
      },
      {
        id: 1, submitter: '王五', location: '公司总部', is_multi_person: false,
        total_amount: 89.00, submit_date: '2026-10-12', status: 'pending',
        remark: '采购打印纸', created_at: '2026-10-12T11:42:56+08:00',
        persons: [
          { id: 1, person_name: '王五', amount: 89, approval_number: 'SP20261012001', status: 'pending' }
        ],
        reasons: [{ id: 1, reason: '办公用品' }]
      }
    ]
  }
}

const mockStats = {
  code: 0,
  data: { total_records: 4, total_amount: 2325.50, pending_count: 2, pending_amount: 1289.00, received_count: 2, received_amount: 1036.50 }
}

const detail4 = mockList.data.list[0]

async function main() {
  const browser = await chromium.launch()
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  page.on('pageerror', e => console.log('[pageerror]', e.message))

  // mock 所有 /api 请求（注意不能拦截 /src/api/index.js，只匹配路径以 /api/ 开头的请求）
  await page.route(url => url.pathname.startsWith('/api/'), route => {
    const url = route.request().url()
    const method = route.request().method()
    let body = { code: 0, data: null }
    if (url.includes('/api/statistics')) body = mockStats
    else if (url.includes('/api/reimbursements/4') && method === 'GET') body = { code: 0, data: detail4 }
    else if (url.includes('/api/reimbursements') && method === 'GET') body = mockList
    else if (url.includes('/api/persons/names')) body = { code: 0, data: ['张三', '李四', '王五', '赵六'] }
    else if (url.includes('/api/reasons/options')) body = { code: 0, data: ['交通', 'ETC', '油费', '误餐', '住宿', '办公用品', '其他'] }
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) })
  })

  // 1. 列表页（等待表格行渲染出来再截）
  await page.goto('http://localhost:5173/', { waitUntil: 'load' })
  await page.waitForSelector('table tbody tr', { timeout: 15000 })
  await page.waitForTimeout(1500)
  await page.screenshot({ path: OUT + '/01-list.png', fullPage: false })
  console.log('01-list.png done')

  // 2. 新增报销弹窗（单人）
  await page.click('button:has-text("新增报销记录")')
  await page.waitForSelector('.el-dialog', { timeout: 10000 })
  await page.waitForTimeout(800)
  await page.screenshot({ path: OUT + '/02-create-single.png' })
  console.log('02-create-single.png done')

  // 3. 切换多人模式截图
  await page.click('.el-dialog .el-radio-button:has-text("多人"), .el-dialog label:has-text("多人")').catch(() => {})
  await page.waitForTimeout(600)
  await page.screenshot({ path: OUT + '/03-create-multi.png' })
  console.log('03-create-multi.png done')
  await page.keyboard.press('Escape')
  await page.waitForTimeout(500)

  // 4. 详情弹窗
  await page.click('table tbody tr:first-child button:has-text("操作")').catch(async () => {
    await page.click('table tbody tr:first-child .el-dropdown').catch(() => {})
  })
  await page.waitForTimeout(600)
  await page.click('.el-dropdown-menu__item:has-text("详情")').catch(() => {})
  await page.waitForSelector('.el-dialog', { timeout: 10000 }).catch(() => {})
  await page.waitForTimeout(800)
  await page.screenshot({ path: OUT + '/04-detail.png' })
  console.log('04-detail.png done')
  await page.keyboard.press('Escape')
  await page.waitForTimeout(500)

  // 5. 操作下拉菜单展开（详情/编辑/标记到账/删除）
  await page.click('table tbody tr:first-child button:has-text("操作")').catch(() => {})
  await page.waitForTimeout(600)
  await page.screenshot({ path: OUT + '/05-actions.png' })
  console.log('05-actions.png done')
  await page.keyboard.press('Escape')
  await page.waitForTimeout(400)

  // 6. 报销缘由多选下拉（在新增弹窗中打开缘由选择器）
  await page.click('button:has-text("新增报销记录")')
  await page.waitForSelector('.el-dialog', { timeout: 10000 })
  await page.waitForTimeout(600)
  await page.click('.el-dialog .el-select').catch(() => {})
  await page.waitForTimeout(800)
  await page.screenshot({ path: OUT + '/06-reasons.png' })
  console.log('06-reasons.png done')

  // 7. 报销人自动补全（输入"张"触发历史姓名联想）
  await page.keyboard.press('Escape')
  await page.waitForTimeout(400)
  await page.fill('.el-dialog input[placeholder*="报销人"]', '张').catch(() => {})
  await page.waitForTimeout(900)
  await page.screenshot({ path: OUT + '/07-autocomplete.png' })
  console.log('07-autocomplete.png done')

  await browser.close()
  console.log('ALL DONE')
}

main().catch(e => { console.error(e); process.exit(1) })
