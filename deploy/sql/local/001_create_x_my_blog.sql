-- 本地 Go 开发库（一次性，需 MySQL 管理员执行）
-- 执行后运行：go run scripts/bootstrap_x_my_blog.go

CREATE DATABASE IF NOT EXISTS `x_my_blog`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_general_ci;

GRANT ALL PRIVILEGES ON `x_my_blog`.* TO 'jiangxia'@'localhost';
FLUSH PRIVILEGES;
