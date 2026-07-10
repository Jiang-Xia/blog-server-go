package tools

// LLMToolCall OpenAI function calling 返回的单条 tool_call。
type LLMToolCall struct {
	Name      string
	Arguments string
}

// ragToolDefinitions OpenAI tools schema（对齐 Nest RAG_TOOL_DEFINITIONS 子集）。
var ragToolDefinitions = []map[string]interface{}{
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_article_ranking",
			"description": "查询全站文章排行榜（浏览/点赞/评论/收藏）",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"metric": map[string]interface{}{"type": "string", "enum": []string{"views", "likes", "comments", "collects"}},
					"limit":  map[string]interface{}{"type": "integer"},
				},
				"required": []string{"metric"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "search_articles",
			"description": "按关键词搜索文章标题或描述",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "string"},
					"limit":   map[string]interface{}{"type": "integer"},
				},
				"required": []string{"keyword"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "list_articles_by_category",
			"description": "按分类名列出文章",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"categoryName": map[string]interface{}{"type": "string"},
					"limit":        map[string]interface{}{"type": "integer"},
				},
				"required": []string{"categoryName"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "list_articles_by_tag",
			"description": "按标签名列出文章",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tagName": map[string]interface{}{"type": "string"},
					"limit":   map[string]interface{}{"type": "integer"},
				},
				"required": []string{"tagName"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_recent_articles",
			"description": "最近发布的文章",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_masterpiece_articles",
			"description": "神作或高等级文章",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_article_stats",
			"description": "单篇文章统计（articleId 或 title）",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"articleId": map[string]interface{}{"type": "integer"},
					"title":     map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_rpg_leaderboard",
			"description": "RPG 排行榜（经验/等级/声望/钻石/签到）",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type":   map[string]interface{}{"type": "string", "enum": []string{"exp", "level", "reputation", "currency", "signDays"}},
					"period": map[string]interface{}{"type": "string", "enum": []string{"total", "week", "month", "season"}},
					"limit":  map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_my_rpg_status",
			"description": "当前登录用户 RPG 状态",
			"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_my_article_stats",
			"description": "当前登录用户发文统计",
			"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "list_authors",
			"description": "列出有已发布文章的作者",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page":     map[string]interface{}{"type": "integer"},
					"pageSize": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_author_stats",
			"description": "查询某作者发文统计",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uid":      map[string]interface{}{"type": "integer"},
					"nickname": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_category_stats",
			"description": "各分类文章数",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_tag_cloud",
			"description": "热门标签云",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_site_nav",
			"description": "站点导航与工具箱入口",
			"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "search_site_pages",
			"description": "在站点页面/工具中搜索",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{"type": "string"},
					"limit":   map[string]interface{}{"type": "integer"},
				},
				"required": []string{"keyword"},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_article_archive_stats",
			"description": "按年归档文章数",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "list_friend_links",
			"description": "友情链接列表",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_msgboard_recent",
			"description": "留言板最近留言",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
	},
}
