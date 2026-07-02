package errcode_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
)

func TestErrCodeInterface(t *testing.T) {
	if errcode.Unauthorized.Code() != 401 {
		t.Fatalf("Unauthorized code want 401, got %d", errcode.Unauthorized.Code())
	}
	if errcode.Unauthorized.Message() == "" {
		t.Fatal("Unauthorized message should not be empty")
	}
}

func TestWithMessage(t *testing.T) {
	ec := errcode.WithMessage(errcode.InvalidParam, "缺少字段 %s", "username")
	if ec.Code() != 400 {
		t.Fatalf("code want 400, got %d", ec.Code())
	}
	if ec.Message() != "缺少字段 username" {
		t.Fatalf("unexpected message: %s", ec.Message())
	}
}

func TestBizErrorString(t *testing.T) {
	if errcode.NotFound.Error() == "" {
		t.Fatal("error string should not be empty")
	}
}
