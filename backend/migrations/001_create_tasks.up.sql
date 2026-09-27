-- 001_create_tasks.up.sql (MySQL 8+)
CREATE TABLE IF NOT EXISTS tasks (
  id CHAR(36) NOT NULL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'todo',
  assignee VARCHAR(255) DEFAULT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL DEFAULT NULL,
  CONSTRAINT chk_tasks_status CHECK (status IN ('todo','in_progress','done')),
  KEY idx_tasks_status (status),
  KEY idx_tasks_assignee (assignee),
  KEY idx_tasks_created_at (created_at),
  KEY idx_tasks_deleted_at (deleted_at),
  -- Duplicate active titles are rejected at DB level -> error 1062 -> HTTP 409.
  UNIQUE KEY ux_tasks_title_active (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
