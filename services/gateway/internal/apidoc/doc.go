// Package apidoc Swagger 元信息（swag 入口文件）。
//
// @title           Blog Gateway API
// @version         1.0
// @description     统一 REST 入口 BFF 与反向代理（gateway :8000）；未列出的 /api/v1/* 按前缀代理至 user/blog/rpg 微服务
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     JWT Bearer Token，格式：Bearer {token}
package apidoc
