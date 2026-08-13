package models

import (
	"time"

	"gorm.io/gorm"
)

// Reimbursement 报销记录主表
type Reimbursement struct {
	ID            uint                  `json:"id" gorm:"primaryKey"`
	Submitter     string                `json:"submitter" gorm:"type:varchar(100);not null;comment:报销人"`
	Location      string                `json:"location" gorm:"type:varchar(255);not null;comment:地点"`
	IsMultiPerson bool                 `json:"is_multi_person" gorm:"default:false;comment:是否多人"`
	TotalAmount   float64               `json:"total_amount" gorm:"type:decimal(10,2);not null;comment:总金额"`
	SubmitDate    string                `json:"submit_date" gorm:"type:date;not null;comment:提交日期"`
	Status        string                `json:"status" gorm:"type:varchar(20);default:'pending';comment:状态 pending=待报销 received=已到账"`
	Remark        string                `json:"remark" gorm:"type:text;comment:备注"`
	Persons       []ReimbursementPerson `json:"persons" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"`
	Reasons       []ReimbursementReason `json:"reasons" gorm:"foreignKey:ReimbursementID;constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

// ReimbursementPerson 报销人员分配（多人报销时每人金额和审批单号）
type ReimbursementPerson struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	ReimbursementID uint    `json:"reimbursement_id" gorm:"not null;index"`
	PersonName      string  `json:"person_name" gorm:"type:varchar(100);not null;comment:人员姓名"`
	Amount          float64 `json:"amount" gorm:"type:decimal(10,2);not null;comment:分配金额"`
	ApprovalNumber  string  `json:"approval_number" gorm:"type:varchar(100);comment:审批单号"`
	Status          string  `json:"status" gorm:"type:varchar(20);default:'pending';comment:个人状态 pending=待报销 received=已到账"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ReimbursementReason 报销缘由（可多项）
type ReimbursementReason struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	ReimbursementID uint   `json:"reimbursement_id" gorm:"not null;index"`
	Reason          string `json:"reason" gorm:"type:varchar(50);not null;comment:报销缘由"`
	CreatedAt       time.Time `json:"created_at"`
}

// 预定义报销缘由选项
var ReasonOptions = []string{"交通", "ETC", "油费", "误餐", "住宿", "办公用品", "其他"}

// CreateReimbursementInput 创建报销记录的输入
type CreateReimbursementInput struct {
	Submitter     string                   `json:"submitter" binding:"required"`
	Location      string                   `json:"location" binding:"required"`
	IsMultiPerson bool                     `json:"is_multi_person"`
	TotalAmount   float64                  `json:"total_amount" binding:"required"`
	SubmitDate    string                   `json:"submit_date" binding:"required"`
	Status        string                   `json:"status"`
	Remark        string                   `json:"remark"`
	Persons       []CreatePersonInput      `json:"persons"`
	Reasons       []string                 `json:"reasons"`
}

type CreatePersonInput struct {
	PersonName     string  `json:"person_name" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	ApprovalNumber string  `json:"approval_number"`
	Status         string  `json:"status"`
}

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required,oneof=pending received"`
}

// AutoMigrate 自动迁移表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Reimbursement{}, &ReimbursementPerson{}, &ReimbursementReason{})
}
