// moderation_service 敏感词审核后同步 comment/msgboard/reply 状态。
package service

import (
	"context"
	"strconv"
	"strings"

	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
)

// ModerationService 内容审核状态同步。
type ModerationService struct {
	comments  *blogrepo.CommentRepo
	msgboards *blogrepo.MsgboardRepo
	replies   *blogrepo.ReplyRepo
}

// NewModerationService 构造 ModerationService。
func NewModerationService(comments *blogrepo.CommentRepo, msgboards *blogrepo.MsgboardRepo, replies *blogrepo.ReplyRepo) *ModerationService {
	return &ModerationService{comments: comments, msgboards: msgboards, replies: replies}
}

// UpdateContentModerationStatus 按来源类型更新实体审核状态。
func (s *ModerationService) UpdateContentModerationStatus(ctx context.Context, sourceType, sourceID, reviewStatus string) (bool, error) {
	if sourceID == "" || sourceID == "0" {
		return false, nil
	}
	status := reviewStatus
	if status != "approved" && status != "rejected" {
		return false, nil
	}
	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case "comment":
		if s.comments == nil {
			return false, nil
		}
		if err := s.comments.UpdateStatus(ctx, sourceID, status); err != nil {
			return false, err
		}
		return true, nil
	case "msgboard":
		if s.msgboards == nil {
			return false, nil
		}
		id, err := strconv.Atoi(sourceID)
		if err != nil {
			return false, err
		}
		if err := s.msgboards.UpdateStatus(ctx, id, status); err != nil {
			return false, err
		}
		return true, nil
	case "reply":
		if s.replies == nil {
			return false, nil
		}
		if err := s.replies.UpdateStatus(ctx, sourceID, status); err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, nil
	}
}
