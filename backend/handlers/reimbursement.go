package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"reimburse-system/config"
	"reimburse-system/models"

	"github.com/gin-gonic/gin"
)

// @Summary 创建报销记录
// @Description 创建一条新的报销记录，支持多人分配和多项报销缘由
// @Tags 报销记录
// @Accept json
// @Produce json
func CreateReimbursement(c *gin.Context) {
	var input models.CreateReimbursementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 默认状态
	if input.Status == "" {
		input.Status = "pending"
	}

	// 如果没有传人员信息，自动用报销人创建一条
	if len(input.Persons) == 0 {
		input.Persons = []models.CreatePersonInput{
			{
				PersonName: input.Submitter,
				Amount:     input.TotalAmount,
				Status:     input.Status,
			},
		}
	}

	// 给人员状态赋默认值
	for i := range input.Persons {
		if input.Persons[i].Status == "" {
			input.Persons[i].Status = "pending"
		}
	}

	// 开启事务
	tx := config.DB.Begin()

	reimb := models.Reimbursement{
		Submitter:     input.Submitter,
		Location:      input.Location,
		IsMultiPerson: input.IsMultiPerson,
		TotalAmount:   input.TotalAmount,
		SubmitDate:    input.SubmitDate,
		Status:        input.Status,
		Remark:        input.Remark,
	}

	if err := tx.Create(&reimb).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	// 创建人员分配
	for _, p := range input.Persons {
		person := models.ReimbursementPerson{
			ReimbursementID: reimb.ID,
			PersonName:      p.PersonName,
			Amount:          p.Amount,
			ApprovalNumber:  p.ApprovalNumber,
			Status:          p.Status,
		}
		if err := tx.Create(&person).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建人员分配失败: " + err.Error()})
			return
		}
	}

	// 创建报销缘由
	for _, r := range input.Reasons {
		reason := models.ReimbursementReason{
			ReimbursementID: reimb.ID,
			Reason:          r,
		}
		if err := tx.Create(&reason).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建报销缘由失败: " + err.Error()})
			return
		}
	}

	tx.Commit()

	// 查询完整数据返回
	config.DB.Preload("Persons").Preload("Reasons").First(&reimb, reimb.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建成功", "data": reimb})
}

// @Summary 获取报销记录列表
// @Description 分页获取报销记录列表，支持按报销人、状态、日期筛选
// @Tags 报销记录
func ListReimbursements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := config.DB.Model(&models.Reimbursement{})

	// 筛选条件
	if submitter := c.Query("submitter"); submitter != "" {
		query = query.Where("submitter LIKE ?", "%"+submitter+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("submit_date >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("submit_date <= ?", endDate)
	}
	if isMulti := c.Query("is_multi_person"); isMulti != "" {
		query = query.Where("is_multi_person = ?", isMulti == "true")
	}

	var total int64
	query.Count(&total)

	var records []models.Reimbursement
	if err := query.Preload("Persons").Preload("Reasons").
		Order("submit_date DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      records,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// @Summary 获取报销记录详情
// @Tags 报销记录
func GetReimbursement(c *gin.Context) {
	id := c.Param("id")
	var record models.Reimbursement
	if err := config.DB.Preload("Persons").Preload("Reasons").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": record})
}

// @Summary 更新报销记录
// @Tags 报销记录
func UpdateReimbursement(c *gin.Context) {
	id := c.Param("id")
	var record models.Reimbursement
	if err := config.DB.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}

	var input models.CreateReimbursementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	tx := config.DB.Begin()

	// 更新主记录
	record.Submitter = input.Submitter
	record.Location = input.Location
	record.IsMultiPerson = input.IsMultiPerson
	record.TotalAmount = input.TotalAmount
	record.SubmitDate = input.SubmitDate
	record.Remark = input.Remark
	if input.Status != "" {
		record.Status = input.Status
	}

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	// 删除旧的子记录
	tx.Where("reimbursement_id = ?", record.ID).Delete(&models.ReimbursementPerson{})
	tx.Where("reimbursement_id = ?", record.ID).Delete(&models.ReimbursementReason{})

	// 重新创建人员分配
	for _, p := range input.Persons {
		person := models.ReimbursementPerson{
			ReimbursementID: record.ID,
			PersonName:      p.PersonName,
			Amount:          p.Amount,
			ApprovalNumber:  p.ApprovalNumber,
			Status:          p.Status,
		}
		if person.Status == "" {
			person.Status = "pending"
		}
		if err := tx.Create(&person).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新人员分配失败: " + err.Error()})
			return
		}
	}

	// 重新创建报销缘由
	for _, r := range input.Reasons {
		reason := models.ReimbursementReason{
			ReimbursementID: record.ID,
			Reason:          r,
		}
		if err := tx.Create(&reason).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新报销缘由失败: " + err.Error()})
			return
		}
	}

	tx.Commit()

	// 查询完整数据返回
	config.DB.Preload("Persons").Preload("Reasons").First(&record, record.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功", "data": record})
}

// @Summary 删除报销记录
// @Tags 报销记录
func DeleteReimbursement(c *gin.Context) {
	id := c.Param("id")
	var record models.Reimbursement
	if err := config.DB.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}

	tx := config.DB.Begin()
	tx.Where("reimbursement_id = ?", record.ID).Delete(&models.ReimbursementPerson{})
	tx.Where("reimbursement_id = ?", record.ID).Delete(&models.ReimbursementReason{})
	tx.Delete(&record)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// @Summary 更新报销状态
// @Tags 报销记录
func UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var record models.Reimbursement
	if err := config.DB.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}

	var input models.UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	tx := config.DB.Begin()
	record.Status = input.Status
	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	// 同时更新所有人员的状态
	if err := tx.Model(&models.ReimbursementPerson{}).
		Where("reimbursement_id = ?", record.ID).
		Update("status", input.Status).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新人员状态失败"})
		return
	}

	tx.Commit()
	config.DB.Preload("Persons").Preload("Reasons").First(&record, record.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "状态更新成功", "data": record})
}

// @Summary 更新单个人员的报销状态
// @Tags 报销记录
func UpdatePersonStatus(c *gin.Context) {
	personID := c.Param("person_id")
	var person models.ReimbursementPerson
	if err := config.DB.First(&person, personID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "人员记录不存在"})
		return
	}

	var input struct {
		Status         string `json:"status" binding:"required,oneof=pending received"`
		ApprovalNumber string `json:"approval_number"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	person.Status = input.Status
	if input.ApprovalNumber != "" {
		person.ApprovalNumber = input.ApprovalNumber
	}
	config.DB.Save(&person)

	// 检查是否所有人员都已到账，如果是则更新主记录状态
	var pendingCount int64
	config.DB.Model(&models.ReimbursementPerson{}).
		Where("reimbursement_id = ? AND status = ?", person.ReimbursementID, "pending").
		Count(&pendingCount)

	if pendingCount == 0 {
		config.DB.Model(&models.Reimbursement{}).
			Where("id = ?", person.ReimbursementID).
			Update("status", "received")
	} else {
		config.DB.Model(&models.Reimbursement{}).
			Where("id = ?", person.ReimbursementID).
			Update("status", "pending")
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "状态更新成功", "data": person})
}

// @Summary 获取所有报销人姓名（去重）
// @Tags 报销记录
func GetPersonNames(c *gin.Context) {
	var names []string
	config.DB.Model(&models.ReimbursementPerson{}).
		Distinct("person_name").
		Pluck("person_name", &names)

	// 也从主记录的 submitter 获取
	var submitters []string
	config.DB.Model(&models.Reimbursement{}).
		Distinct("submitter").
		Pluck("submitter", &submitters)

	nameSet := make(map[string]bool)
	var result []string
	for _, n := range names {
		if !nameSet[n] {
			nameSet[n] = true
			result = append(result, n)
		}
	}
	for _, n := range submitters {
		if !nameSet[n] {
			nameSet[n] = true
			result = append(result, n)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// @Summary 获取报销缘由选项
// @Tags 报销记录
func GetReasonOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": models.ReasonOptions,
	})
}

// @Summary 获取统计数据
// @Tags 报销记录
func GetStatistics(c *gin.Context) {
	var totalRecords int64
	config.DB.Model(&models.Reimbursement{}).Count(&totalRecords)

	var pendingCount int64
	config.DB.Model(&models.Reimbursement{}).Where("status = ?", "pending").Count(&pendingCount)

	var receivedCount int64
	config.DB.Model(&models.Reimbursement{}).Where("status = ?", "received").Count(&receivedCount)

	type AmountResult struct {
		Total float64
	}
	var pendingAmount AmountResult
	config.DB.Model(&models.Reimbursement{}).
		Where("status = ?", "pending").
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&pendingAmount)

	var receivedAmount AmountResult
	config.DB.Model(&models.Reimbursement{}).
		Where("status = ?", "received").
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&receivedAmount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total_records":   totalRecords,
			"pending_count":   pendingCount,
			"received_count":  receivedCount,
			"pending_amount":  pendingAmount.Total,
			"received_amount": receivedAmount.Total,
		},
	})
}

// 搜索人员姓名辅助
func searchPersons(keyword string) []string {
	var names []string
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return names
	}
	config.DB.Model(&models.ReimbursementPerson{}).
		Distinct("person_name").
		Where("person_name LIKE ?", "%"+keyword+"%").
		Pluck("person_name", &names)
	return names
}
