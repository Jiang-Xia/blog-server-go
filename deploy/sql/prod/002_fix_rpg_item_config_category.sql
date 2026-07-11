-- 修复 x_my_blog 导入备份时误将 category 列命名为 x_category（对齐 Nest rpg_item_config）
-- 幂等：仅当存在 x_category 且不存在 category 时执行

SET @has_x_category := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'x_rpg_item_config'
    AND COLUMN_NAME = 'x_category'
);
SET @has_category := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'x_rpg_item_config'
    AND COLUMN_NAME = 'category'
);

SET @sql := IF(
  @has_x_category > 0 AND @has_category = 0,
  'ALTER TABLE `x_rpg_item_config` CHANGE COLUMN `x_category` `category` varchar(64) NOT NULL DEFAULT '''' COMMENT ''分类''',
  'SELECT ''skip: x_rpg_item_config.category already aligned'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
