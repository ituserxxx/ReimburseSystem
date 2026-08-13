-- ============================================================
-- 报销记录系统 数据库初始化脚本
-- 适用：MySQL 5.7+ / 8.0+
-- 用法：
--   1) 先确保已创建数据库（见下方 CREATE DATABASE）
--   2) 在本文件对应数据库下执行即可
-- 说明：
--   本脚本会先 DROP 三张表再重建，适合「干净初始化 / 重新建表」。
--   若只想追加缺失字段而不清空数据，请改用后端 AutoMigrate（main.go 已内置自愈迁移）。
-- ============================================================

CREATE DATABASE IF NOT EXISTS `reimburse_system` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `reimburse_system`;

-- 关闭外键检查，便于按任意顺序 DROP
SET FOREIGN_KEY_CHECKS = 0;

-- 报销记录主表
DROP TABLE IF EXISTS `reimbursements`;
CREATE TABLE `reimbursements` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `submitter` VARCHAR(100) NOT NULL COMMENT '报销人',
  `location` VARCHAR(255) NOT NULL COMMENT '地点（必填）',
  `is_multi_person` TINYINT(1) DEFAULT 0 COMMENT '是否多人: 0=单人 1=多人',
  `total_amount` DECIMAL(10,2) NOT NULL COMMENT '总金额',
  `submit_date` DATE NOT NULL COMMENT '提交日期',
  `status` VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending=待报销 received=已到账',
  `remark` TEXT COMMENT '备注',
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX `idx_submitter` (`submitter`),
  INDEX `idx_status` (`status`),
  INDEX `idx_submit_date` (`submit_date`),
  INDEX `idx_location` (`location`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报销记录主表';

-- 报销人员分配表（多人报销时每人金额和审批单号；单人报销固定 1 条）
DROP TABLE IF EXISTS `reimbursement_persons`;
CREATE TABLE `reimbursement_persons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `reimbursement_id` BIGINT UNSIGNED NOT NULL COMMENT '报销记录ID',
  `person_name` VARCHAR(100) NOT NULL COMMENT '人员姓名',
  `amount` DECIMAL(10,2) NOT NULL COMMENT '分配金额',
  `approval_number` VARCHAR(100) DEFAULT '' COMMENT '审批单号（按人员对应）',
  `status` VARCHAR(20) DEFAULT 'pending' COMMENT '个人状态: pending=待报销 received=已到账',
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX `idx_reimbursement_id` (`reimbursement_id`),
  INDEX `idx_person_name` (`person_name`),
  CONSTRAINT `fk_person_reimbursement` FOREIGN KEY (`reimbursement_id`) REFERENCES `reimbursements`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报销人员分配表';

-- 报销缘由表（可多项，每条记录对应多个缘由）
DROP TABLE IF EXISTS `reimbursement_reasons`;
CREATE TABLE `reimbursement_reasons` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `reimbursement_id` BIGINT UNSIGNED NOT NULL COMMENT '报销记录ID',
  `reason` VARCHAR(50) NOT NULL COMMENT '报销缘由',
  `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX `idx_reimbursement_id` (`reimbursement_id`),
  CONSTRAINT `fk_reason_reimbursement` FOREIGN KEY (`reimbursement_id`) REFERENCES `reimbursements`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报销缘由表';

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 取值约定（与后端 models.ReasonOptions / Status 保持一致）
-- ============================================================
-- 报销缘由 reason 可选值：
--   交通 / ETC / 油费 / 误餐 / 住宿 / 办公用品 / 其他
-- 主表 status 取值：
--   pending = 待报销
--   received = 已到账
-- 人员表 status 取值（与主表一致）：
--   pending = 待报销
--   received = 已到账
--
-- 注意：后端使用 GORM AutoMigrate 会自动创建/补齐表结构（含 location 自愈），
-- 此脚本主要用于手动初始化、重新建表或了解表结构。
