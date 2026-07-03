package tools

// NavLink 站点导航入口。
type NavLink struct {
	Path  string
	Title string
}

var siteNavLinks = []NavLink{
	{Path: "/", Title: "首页"},
	{Path: "/rpg", Title: "冒险"},
	{Path: "/explore", Title: "快速入口"},
	{Path: "/archives", Title: "归档"},
	{Path: "/links", Title: "友链"},
	{Path: "/msgboard", Title: "留言板"},
	{Path: "/about", Title: "关于"},
	{Path: "/projects", Title: "项目"},
	{Path: "/tool", Title: "工具箱"},
	{Path: "/features", Title: "特性"},
	{Path: "/features/rpg-guide", Title: "RPG 冒险攻略"},
}

var toolLinks = []NavLink{
	{Path: "/tool/codes", Title: "条形/二维码"},
	{Path: "/tool/pdf", Title: "PDF"},
	{Path: "/tool/watermark", Title: "水印"},
	{Path: "/tool/photos", Title: "光影边框"},
	{Path: "/tool/audio-visualized", Title: "音频可视化"},
	{Path: "/tool/upload-slice", Title: "切片上传"},
	{Path: "/tool/other", Title: "其他工具"},
	{Path: "/tool/webrtc", Title: "WebRTC"},
	{Path: "/tool/test", Title: "测试"},
	{Path: "/tool/rsa", Title: "RSA加解密工具"},
	{Path: "/tool/des", Title: "对称加密工具"},
	{Path: "/tool/sm", Title: "国密加密工具"},
	{Path: "/tool/ai", Title: "AI"},
	{Path: "/tool/ai-summary", Title: "AI文章摘要"},
}

// featurePages 特性页导航（与 rag/static_page.go RAGStaticPages 保持同步）。
var featurePages = []struct {
	Title, URL, Description string
}{
	{Title: "站点特性概览", URL: "/features", Description: "博客核心功能模块与 RPG 冒险体系概览。"},
	{Title: "博客 RPG 冒险攻略", URL: "/features/rpg-guide", Description: "从签到升级到赛季排行的完整 RPG 玩法攻略。"},
	{Title: "工具箱说明", URL: "/tool", Description: "站内 14+ 在线工具的用途与入口路径说明。"},
}
