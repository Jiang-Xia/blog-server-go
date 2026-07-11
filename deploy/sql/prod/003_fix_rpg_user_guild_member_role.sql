-- 修复 x_my_blog 导入备份时误将 role 列命名为 x_role（与 x_role 权限表冲突）
-- 幂等：仅当存在 x_role 且不存在 role 时执行

SET @has_x_role := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'x_rpg_user_guild_member'
    AND COLUMN_NAME = 'x_role'
);
SET @has_role := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'x_rpg_user_guild_member'
    AND COLUMN_NAME = 'role'
);

SET @sql := IF(
  @has_x_role > 0 AND @has_role = 0,
  'ALTER TABLE `x_rpg_user_guild_member` CHANGE COLUMN `x_role` `role` varchar(32) NOT NULL DEFAULT ''member'' COMMENT ''角色: leader/officer/member''',
  'SELECT ''skip: x_rpg_user_guild_member.role already aligned'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
