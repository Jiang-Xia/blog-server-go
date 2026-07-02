-- 生产：新建 Go 专用库 x_my_blog（保留 Nest myblog 不动）
-- 必须以 MySQL root 执行：
--   sudo mysql -u root -p < deploy/sql/prod/001_create_x_my_blog.sql
--
-- 本脚本只建库。授权见下方（jxblog 账户 host 可能是 127.0.0.1 而非 localhost）。

CREATE DATABASE IF NOT EXISTS `x_my_blog`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_general_ci;
