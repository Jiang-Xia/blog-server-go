// Package ent 是 Ent ORM 生成代码根包；schema 定义在 schema/ 子目录，执行 make ent-gen 重新生成客户端。
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema --target ./ --feature sql/upsert
