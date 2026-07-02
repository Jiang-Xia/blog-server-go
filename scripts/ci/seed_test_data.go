// 写入 CI 测试超级管理员（18888888888 / super，roleId=1）。
//
//	go run scripts/ci/seed_test_data.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
	userent "github.com/Jiang-Xia/blog-server-go/services/user/ent"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent/role"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent/user"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/Jiang-Xia/blog-server-go/services/user/ent/runtime"
)

const (
	seedUsername = "18888888888"
	seedPassword = "super"
)

func main() {
	ctx := context.Background()
	client, err := userent.Open("mysql", mysqlDSN())
	if err != nil {
		fail(err)
	}
	defer client.Close()

	hash, err := crypto.Hash(seedPassword)
	if err != nil {
		fail(err)
	}

	if _, err := client.Role.Get(ctx, 1); userent.IsNotFound(err) {
		if _, err := client.Role.Create().
			SetID(1).
			SetRoleName("super").
			SetRoleDesc("CI 超级管理员").
			Save(ctx); err != nil {
			fail(fmt.Errorf("create role: %w", err))
		}
		fmt.Println("seed role id=1")
	}

	exists, err := client.User.Query().Where(user.UsernameEQ(seedUsername)).Exist(ctx)
	if err != nil {
		fail(err)
	}
	if exists {
		fmt.Println("seed user already exists")
		return
	}

	if _, err := client.User.Create().
		SetID(1).
		SetUsername(seedUsername).
		SetNickname("super").
		SetPassword(hash).
		SetSalt("").
		SetStatus("active").
		AddRoleIDs(1).
		Save(ctx); err != nil {
		fail(fmt.Errorf("create user: %w", err))
	}
	fmt.Println("seed user", seedUsername)

	// 确认角色关联
	n, err := client.User.Query().Where(user.UsernameEQ(seedUsername)).QueryRoles().Where(role.IDEQ(1)).Count(ctx)
	if err != nil || n == 0 {
		fail(fmt.Errorf("user role link missing"))
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "seed_test_data: %v\n", err)
	os.Exit(1)
}
