-- MySQL dump 10.13  Distrib 9.0.1, for Win64 (x86_64)
--
-- Host: 127.0.0.1    Database: x_my_blog
-- ------------------------------------------------------
-- Server version	9.0.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `x_article`
--

DROP TABLE IF EXISTS `x_article`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_article` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `uTime` longtext COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint NOT NULL,
  `deptId` bigint DEFAULT NULL,
  `isDelete` tinyint(1) NOT NULL DEFAULT '0',
  `topping` bigint NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `title` longtext COLLATE utf8mb4_bin NOT NULL,
  `cover` longtext COLLATE utf8mb4_bin NOT NULL,
  `likes` bigint NOT NULL DEFAULT '0',
  `views` bigint NOT NULL DEFAULT '0',
  `articleExp` bigint NOT NULL DEFAULT '0',
  `articleLevel` bigint NOT NULL DEFAULT '1',
  `reputationGained` bigint NOT NULL DEFAULT '0',
  `isMasterpiece` bigint NOT NULL DEFAULT '0',
  `tipTotal` bigint NOT NULL DEFAULT '0',
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'publish',
  `description` longtext COLLATE utf8mb4_bin NOT NULL,
  `contentHtml` longtext COLLATE utf8mb4_bin NOT NULL,
  `useArticles` bigint DEFAULT NULL,
  `articles` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `content` longtext COLLATE utf8mb4_bin NOT NULL,
  `scheduledPublishAt` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_61ae80cb2a104883f07da6b90e9` (`useArticles`),
  KEY `FK_0845fb04306c6dacf1c847cb953` (`articles`),
  KEY `FK_6f973802c779bbf4ee90eab8059` (`deptId`)
) ENGINE=InnoDB AUTO_INCREMENT=153 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_article_tags_tag`
--

DROP TABLE IF EXISTS `x_article_tags_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_article_tags_tag` (
  `articleId` bigint NOT NULL,
  `tagId` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `IDX_9b7dd28292e2799512cd70bfd8` (`articleId`),
  KEY `IDX_5fee2a10f8d6688bd2f2c50f15` (`tagId`)
) ENGINE=InnoDB AUTO_INCREMENT=231 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_category`
--

DROP TABLE IF EXISTS `x_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_category` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint NOT NULL,
  `label` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `value` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `color` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `create_at` timestamp NOT NULL,
  `update_at` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_collect`
--

DROP TABLE IF EXISTS `x_collect`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_collect` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint NOT NULL,
  `createTime` timestamp NOT NULL,
  `articleId` bigint NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_54569140f8529db8c856904c726` (`articleId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_comment`
--

DROP TABLE IF EXISTS `x_comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_comment` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `content` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint NOT NULL,
  `userId` bigint DEFAULT NULL,
  `articleId` bigint DEFAULT NULL,
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'approved',
  PRIMARY KEY (`id`),
  KEY `FK_c0354a9a009d3bb45a08655ce3b` (`userId`),
  KEY `FK_c20404221e5c125a581a0d90c0e` (`articleId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_dept`
--

DROP TABLE IF EXISTS `x_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_dept` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `deptName` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `deptCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `parentId` bigint NOT NULL DEFAULT '0',
  `leaderId` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `leaderName` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `orderNum` bigint NOT NULL DEFAULT '0',
  `status` bigint NOT NULL DEFAULT '1',
  `remark` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `isDelete` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_d71a8c6d61bfbada1b9a8e6132` (`deptCode`),
  UNIQUE KEY `deptCode` (`deptCode`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_file`
--

DROP TABLE IF EXISTS `x_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_file` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `pid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '0',
  `isFolder` bigint NOT NULL DEFAULT '0',
  `originalname` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `filename` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `type` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `size` bigint NOT NULL,
  `url` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `create_at` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_knowledge_chunk`
--

DROP TABLE IF EXISTS `x_knowledge_chunk`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_knowledge_chunk` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `article_id` int NOT NULL,
  `chunk_index` int NOT NULL,
  `title` varchar(512) COLLATE utf8mb4_general_ci NOT NULL,
  `content` text COLLATE utf8mb4_general_ci NOT NULL,
  `url` varchar(512) COLLATE utf8mb4_general_ci NOT NULL,
  `category` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `tags` json DEFAULT NULL,
  `embedding_json` json NOT NULL,
  `status` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'active',
  `indexed_at` datetime(6) NOT NULL,
  `create_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `update_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `source_type` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'article',
  `source_key` varchar(128) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `heading_path` varchar(512) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `content_type` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'prose',
  `search_text` text COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8137 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_like`
--

DROP TABLE IF EXISTS `x_like`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_like` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `articleId` bigint NOT NULL,
  `uid` bigint NOT NULL DEFAULT '-999',
  `ip` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_a95ce350aee91167d8a1cefeb97` (`articleId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_link`
--

DROP TABLE IF EXISTS `x_link`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_link` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `icon` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `url` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `title` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `desp` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `agreed` bigint NOT NULL DEFAULT '0',
  `lastCheckStatus` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'unchecked',
  `lastCheckTime` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_menu`
--

DROP TABLE IF EXISTS `x_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_menu` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `pid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '0',
  `path` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `order` bigint NOT NULL DEFAULT '1',
  `icon` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `locale` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `requiresAuth` bigint NOT NULL DEFAULT '1',
  `filePath` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `isDelete` tinyint NOT NULL DEFAULT '0',
  `menuCnName` varchar(255) COLLATE utf8mb4_bin DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_msgboard`
--

DROP TABLE IF EXISTS `x_msgboard`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_msgboard` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `eamil` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `address` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `comment` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `avatar` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `location` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `system` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `browser` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `respondent` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `imgUrl` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `ip` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `pId` bigint NOT NULL DEFAULT '0',
  `replyId` bigint DEFAULT NULL,
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'approved',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1372 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_my_file`
--

DROP TABLE IF EXISTS `x_my_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_my_file` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_operation_log`
--

DROP TABLE IF EXISTS `x_operation_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_operation_log` (
  `createTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `userId` bigint NOT NULL,
  `username` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `module` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `action` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `method` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `path` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `ip` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `requestBody` longtext COLLATE utf8mb4_bin,
  `statusCode` bigint NOT NULL DEFAULT '200',
  PRIMARY KEY (`id`),
  KEY `IDX_0aed8f911622e7e2c69f6bda19` (`userId`,`module`)
) ENGINE=InnoDB AUTO_INCREMENT=1716 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_pay_order`
--

DROP TABLE IF EXISTS `x_pay_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_pay_order` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `outTradeNo` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `tradeNo` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `subject` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `totalAmount` double NOT NULL DEFAULT '0',
  `buyerOpenId` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'PENDING',
  `refundAmount` double NOT NULL DEFAULT '0',
  `channel` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'alipay',
  `extendParams` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_0ab64c03d023a67c281c3389ec` (`outTradeNo`),
  UNIQUE KEY `outTradeNo` (`outTradeNo`)
) ENGINE=InnoDB AUTO_INCREMENT=93 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_privilege`
--

DROP TABLE IF EXISTS `x_privilege`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_privilege` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `privilegeName` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `privilegeCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `privilegePage` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `isVisible` bigint NOT NULL DEFAULT '1',
  `pathPattern` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `httpMethod` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `isPublic` bigint NOT NULL DEFAULT '0',
  `requireOwnership` bigint NOT NULL DEFAULT '0',
  `description` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `isDelete` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=281 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rag_index_job`
--

DROP TABLE IF EXISTS `x_rag_index_job`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rag_index_job` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `article_id` int NOT NULL DEFAULT '0',
  `status` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'pending',
  `chunk_count` int NOT NULL DEFAULT '0',
  `error_msg` text COLLATE utf8mb4_general_ci,
  `create_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `update_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rag_query_log`
--

DROP TABLE IF EXISTS `x_rag_query_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rag_query_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` int NOT NULL,
  `question` varchar(500) COLLATE utf8mb4_general_ci NOT NULL,
  `answer_preview` varchar(1000) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `citations_json` json DEFAULT NULL,
  `latency_ms` int NOT NULL DEFAULT '0',
  `status` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'success',
  `create_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_reply`
--

DROP TABLE IF EXISTS `x_reply`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_reply` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `parentId` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `replyUid` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `content` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint NOT NULL,
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'approved',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_role`
--

DROP TABLE IF EXISTS `x_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_role` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `roleName` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `roleDesc` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `isDelete` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_role_data_scope`
--

DROP TABLE IF EXISTS `x_role_data_scope`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_role_data_scope` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `roleId` bigint NOT NULL,
  `resourceType` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'article',
  `scopeType` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `deptIds` json DEFAULT NULL,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `isDelete` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `FK_be62aa51e9cb1bb1afb26fa7da5` (`roleId`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_role_menus_menu`
--

DROP TABLE IF EXISTS `x_role_menus_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_role_menus_menu` (
  `roleId` bigint NOT NULL,
  `menuId` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `IDX_eec9c5cb17157b2294fd9f0edb` (`roleId`),
  KEY `IDX_f1adc6be166630ee2476d7bbf0` (`menuId`)
) ENGINE=InnoDB AUTO_INCREMENT=124 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_role_privileges_privilege`
--

DROP TABLE IF EXISTS `x_role_privileges_privilege`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_role_privileges_privilege` (
  `roleId` bigint NOT NULL,
  `privilegeId` bigint NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `IDX_d11ab7c8589ca17646c5345fb7` (`roleId`),
  KEY `IDX_e04315305e9b12cc7e18bda6ef` (`privilegeId`)
) ENGINE=InnoDB AUTO_INCREMENT=309 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_role_users_user`
--

DROP TABLE IF EXISTS `x_role_users_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_role_users_user` (
  `userId` bigint NOT NULL,
  `roleId` bigint NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `IDX_a88fcb405b56bf2e2646e9d479` (`userId`),
  KEY `IDX_ed6edac7184b013d4bd58d60e5` (`roleId`)
) ENGINE=InnoDB AUTO_INCREMENT=54 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg`
--

DROP TABLE IF EXISTS `x_rpg`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `exp` bigint NOT NULL DEFAULT '0',
  `level` bigint NOT NULL DEFAULT '1',
  `lifeValue` bigint NOT NULL DEFAULT '100',
  `lastSignDate` timestamp NULL DEFAULT NULL,
  `totalSignDays` bigint NOT NULL DEFAULT '0',
  `consecutiveSignDays` bigint NOT NULL DEFAULT '0',
  `banStartTime` timestamp NULL DEFAULT NULL,
  `banEndTime` timestamp NULL DEFAULT NULL,
  `sensitiveHitsCount` bigint NOT NULL DEFAULT '0',
  `zeroLifeCount` bigint NOT NULL DEFAULT '0',
  `lotteryTickets` bigint NOT NULL DEFAULT '0',
  `reputation` bigint NOT NULL DEFAULT '0',
  `lotteryPityCounter` bigint NOT NULL DEFAULT '0',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `lotteryLegendaryPityCounter` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_ac218940269d84be6ed90bc206` (`uid`),
  UNIQUE KEY `REL_ac218940269d84be6ed90bc206` (`uid`),
  UNIQUE KEY `uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=82 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_activity`
--

DROP TABLE IF EXISTS `x_rpg_activity`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_activity` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `code` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `startTime` timestamp NOT NULL,
  `endTime` timestamp NOT NULL,
  `expBuffRate` double NOT NULL DEFAULT '1',
  `active` bigint NOT NULL DEFAULT '1',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `activityType` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'event',
  `posterUrl` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_172a2315ffd5ab8a839f4c441c` (`code`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_article_tip`
--

DROP TABLE IF EXISTS `x_rpg_article_tip`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_article_tip` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `articleId` bigint NOT NULL,
  `authorUid` bigint NOT NULL,
  `amount` bigint NOT NULL,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_guild`
--

DROP TABLE IF EXISTS `x_rpg_guild`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_guild` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `leaderUid` bigint NOT NULL,
  `announcement` longtext COLLATE utf8mb4_bin,
  `memberCount` bigint NOT NULL DEFAULT '1',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_9332f106de5e746a9e43be36d5` (`name`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_item_config`
--

DROP TABLE IF EXISTS `x_rpg_item_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_item_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `code` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `sort` bigint NOT NULL DEFAULT '10',
  `active` bigint NOT NULL DEFAULT '1',
  `isHidden` bigint NOT NULL DEFAULT '0',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `itemType` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `category` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `icon` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'default',
  `rarity` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'common',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_9a10da280b0bda39237230ced9` (`code`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_leaderboard_snapshot`
--

DROP TABLE IF EXISTS `x_rpg_leaderboard_snapshot`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_leaderboard_snapshot` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `score` bigint NOT NULL DEFAULT '0',
  `rank` bigint NOT NULL DEFAULT '0',
  `createTime` timestamp NOT NULL,
  `periodType` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `periodKey` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `scoreType` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_level_reward`
--

DROP TABLE IF EXISTS `x_rpg_level_reward`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_level_reward` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `level` bigint NOT NULL,
  `avatarFrame` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `title` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `currencyReward` bigint NOT NULL DEFAULT '0',
  `active` bigint NOT NULL DEFAULT '1',
  `sort` bigint NOT NULL DEFAULT '10',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_3e82a78cf059257f86c3af1711` (`level`),
  UNIQUE KEY `level` (`level`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_lottery_pool`
--

DROP TABLE IF EXISTS `x_rpg_lottery_pool`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_lottery_pool` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `itemCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `probability` double NOT NULL,
  `active` bigint NOT NULL DEFAULT '1',
  `sort` bigint NOT NULL DEFAULT '10',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `rarity` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_quest`
--

DROP TABLE IF EXISTS `x_rpg_quest`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_quest` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `code` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `type` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'daily',
  `targetAction` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `targetCount` bigint NOT NULL DEFAULT '1',
  `expReward` bigint NOT NULL DEFAULT '10',
  `hpReward` bigint NOT NULL DEFAULT '0',
  `currencyReward` bigint NOT NULL DEFAULT '0',
  `active` bigint NOT NULL DEFAULT '1',
  `sort` bigint NOT NULL DEFAULT '10',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `questSubtype` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'daily',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_0ec384f2a14413a95dfb13cfc8` (`code`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_achievement`
--

DROP TABLE IF EXISTS `x_rpg_user_achievement`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_achievement` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `achievementCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `progress` bigint NOT NULL DEFAULT '0',
  `completed` bigint NOT NULL DEFAULT '0',
  `completedAt` timestamp NULL DEFAULT NULL,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_a402e09c2ca20b1e3819465376` (`uid`,`achievementCode`)
) ENGINE=InnoDB AUTO_INCREMENT=138 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_buff`
--

DROP TABLE IF EXISTS `x_rpg_user_buff`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_buff` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `buffCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `buffType` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `value` double NOT NULL,
  `expireAt` timestamp NOT NULL,
  `remainingUses` bigint NOT NULL DEFAULT '1',
  `isActive` bigint NOT NULL DEFAULT '1',
  `sourceType` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `sourceId` bigint DEFAULT NULL,
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `triggerMode` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'auto',
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=97 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_guild_member`
--

DROP TABLE IF EXISTS `x_rpg_user_guild_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_guild_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `guildId` bigint NOT NULL,
  `uid` bigint NOT NULL,
  `joinTime` timestamp NOT NULL,
  `role` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'member',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_b0bb33d146a93c326963e620e8` (`uid`),
  UNIQUE KEY `IDX_5b9572221e6511d0ea8c932094` (`guildId`,`uid`),
  UNIQUE KEY `uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_inventory`
--

DROP TABLE IF EXISTS `x_rpg_user_inventory`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_inventory` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `itemCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `quantity` bigint NOT NULL DEFAULT '1',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `acquiredAt` timestamp NOT NULL,
  `source` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'system',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_793032084446be59b619cd8734` (`uid`,`itemCode`)
) ENGINE=InnoDB AUTO_INCREMENT=108 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_loadout`
--

DROP TABLE IF EXISTS `x_rpg_user_loadout`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_loadout` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `titleCode` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `avatarFrameCode` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `petId` bigint DEFAULT NULL,
  `effectJson` longtext COLLATE utf8mb4_bin,
  `updateTime` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '最后更换时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_3ca81bdb5cb0eb91dcb6e4e436` (`uid`),
  UNIQUE KEY `uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=83 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_lottery_record`
--

DROP TABLE IF EXISTS `x_rpg_user_lottery_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_lottery_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `poolItemCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `itemName` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `rarity` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=639 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_pet`
--

DROP TABLE IF EXISTS `x_rpg_user_pet`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_pet` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `petCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `level` bigint NOT NULL DEFAULT '1',
  `exp` bigint NOT NULL DEFAULT '0',
  `effectJson` longtext COLLATE utf8mb4_bin,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `nickname` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_quest_progress`
--

DROP TABLE IF EXISTS `x_rpg_user_quest_progress`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_quest_progress` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `questCode` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `progress` bigint NOT NULL DEFAULT '0',
  `completed` bigint NOT NULL DEFAULT '0',
  `claimed` bigint NOT NULL DEFAULT '0',
  `questDate` timestamp NOT NULL,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_8d7bc273eb2f22749d31691630` (`uid`,`questCode`,`questDate`)
) ENGINE=InnoDB AUTO_INCREMENT=177 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_rpg_user_social_log`
--

DROP TABLE IF EXISTS `x_rpg_user_social_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_rpg_user_social_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `fromUid` bigint NOT NULL,
  `toUid` bigint NOT NULL,
  `costCurrency` bigint NOT NULL DEFAULT '0',
  `hpDelta` bigint NOT NULL DEFAULT '0',
  `createTime` timestamp NOT NULL,
  `action` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `updateTime` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_scheduled_task`
--

DROP TABLE IF EXISTS `x_scheduled_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_scheduled_task` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `description` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `cron` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `cronHuman` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `enabled` bigint NOT NULL DEFAULT '1',
  `logRecording` bigint NOT NULL DEFAULT '1',
  `sortOrder` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_d83e0fe52ade6e88fb3558b4a5` (`name`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_scheduled_task_log`
--

DROP TABLE IF EXISTS `x_scheduled_task_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_scheduled_task_log` (
  `createTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `taskName` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `startTime` timestamp NOT NULL,
  `endTime` timestamp NULL DEFAULT NULL,
  `result` longtext COLLATE utf8mb4_bin,
  `errorMessage` longtext COLLATE utf8mb4_bin,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=92 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_sensitive_word`
--

DROP TABLE IF EXISTS `x_sensitive_word`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_sensitive_word` (
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `word` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `category` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '自定义',
  `status` bigint NOT NULL DEFAULT '1',
  `level` bigint NOT NULL DEFAULT '2',
  `hpPenalty` bigint NOT NULL DEFAULT '20',
  `needReview` bigint NOT NULL DEFAULT '1',
  `action` bigint NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_4769a287fc92eb7f26a1f593ef` (`word`),
  UNIQUE KEY `word` (`word`)
) ENGINE=InnoDB AUTO_INCREMENT=600 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_sensitive_word_hit`
--

DROP TABLE IF EXISTS `x_sensitive_word_hit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_sensitive_word_hit` (
  `createTime` timestamp NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `sourceType` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `sourceId` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `content` longtext COLLATE utf8mb4_bin NOT NULL,
  `hitWords` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint DEFAULT NULL,
  `ip` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'pending',
  `reviewerId` bigint DEFAULT NULL,
  `reviewTime` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_site_notification`
--

DROP TABLE IF EXISTS `x_site_notification`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_site_notification` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uid` bigint NOT NULL,
  `type` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `payload` longtext COLLATE utf8mb4_bin NOT NULL,
  `read` bigint NOT NULL DEFAULT '0',
  `createTime` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_tag`
--

DROP TABLE IF EXISTS `x_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_tag` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `uid` bigint NOT NULL,
  `label` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `value` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `color` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `create_at` timestamp NOT NULL,
  `update_at` timestamp NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `x_user`
--

DROP TABLE IF EXISTS `x_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `x_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `createTime` timestamp NOT NULL,
  `updateTime` timestamp NOT NULL,
  `isDelete` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT 'active',
  `password` longtext COLLATE utf8mb4_bin NOT NULL,
  `salt` longtext COLLATE utf8mb4_bin NOT NULL,
  `intro` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `avatar` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `homepage` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `email` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `githubId` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `wechatOpenId` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `username` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `deptId` bigint DEFAULT NULL,
  `nickname` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `IDX_78a916df40e02a9deb1c4b75ed` (`username`),
  UNIQUE KEY `IDX_e12875dfb3b1d92d7d7c5377e2` (`email`),
  UNIQUE KEY `IDX_0d84cc6a830f0e4ebbfcd6381d` (`githubId`),
  UNIQUE KEY `IDX_f90cbd0e30758ea9ff6afbd4cf` (`wechatOpenId`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `githubId` (`githubId`),
  UNIQUE KEY `wechatOpenId` (`wechatOpenId`),
  UNIQUE KEY `username` (`username`),
  KEY `FK_b79e66a3f148e12f9eb5dafb3c0` (`deptId`)
) ENGINE=InnoDB AUTO_INCREMENT=58 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-07-15 15:32:50
