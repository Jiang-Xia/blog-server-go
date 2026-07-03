package userport

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
)

type mockArticleScope struct {
	activeIDs []int
	dept      *usersvc.DeptDTO
	deptIDs   []int
	assertErr error
}

func (m *mockArticleScope) ListActiveUserIDs(_ context.Context) ([]int, error) {
	return m.activeIDs, nil
}

func (m *mockArticleScope) GetDept(_ context.Context, id int) (*usersvc.DeptDTO, error) {
	if m.dept == nil {
		return nil, errors.New("dept not found")
	}
	return m.dept, nil
}

func (m *mockArticleScope) ResolveArticleAccessibleDeptIDs(_ context.Context, _ int) ([]int, error) {
	return m.deptIDs, nil
}

func (m *mockArticleScope) AssertArticleDeptAccess(_ context.Context, _ int, _ *int) error {
	return m.assertErr
}

type mockUserService struct {
	user *usersvc.UserDTO
}

func (m *mockUserService) GetUser(_ context.Context, _ uint64) (*usersvc.UserDTO, error) {
	return m.user, nil
}

func (m *mockUserService) GetUserBatch(_ context.Context, _ []uint64) ([]*usersvc.UserDTO, error) {
	return nil, nil
}

func TestGRPCArticleUserPortListActiveUserIDs(t *testing.T) {
	scope := &mockArticleScope{activeIDs: []int{1, 2}}
	port := &GRPCArticleUserPort{scope: scope, users: &mockUserService{}}
	ids, err := port.ListActiveUserIDs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != 1 {
		t.Fatalf("ids=%v", ids)
	}
}

func TestGRPCArticleAdminPortResolveDeptIDs(t *testing.T) {
	scope := &mockArticleScope{deptIDs: []int{4}}
	port := NewGRPCArticleAdminPort(scope)
	ids, err := port.ResolveArticleAccessibleDeptIDs(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != 4 {
		t.Fatalf("ids=%v", ids)
	}
}

func TestGRPCArticleUserPortFindUserForArticleDept(t *testing.T) {
	deptID := 4
	port := &GRPCArticleUserPort{
		scope: &mockArticleScope{},
		users: &mockUserService{user: &usersvc.UserDTO{ID: 1, Status: "active", DeptID: &deptID}},
	}
	u, err := port.FindUserForArticle(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if u.DeptID == nil || *u.DeptID != 4 {
		t.Fatalf("dept=%v", u.DeptID)
	}
}
