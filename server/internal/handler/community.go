package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
)

type CommunityHandler struct {
	community *service.CommunityService
}

func NewCommunityHandler(community *service.CommunityService) *CommunityHandler {
	return &CommunityHandler{community: community}
}

func (h *CommunityHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	posts, total, normalizedPage, normalizedPageSize, err := h.community.ListPosts(c.Request.Context(), page, pageSize, c.Query("tag"))
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.PostResponse, 0, len(posts))
	for _, post := range posts {
		postCopy := post
		items = append(items, dto.ToPostResponse(&postCopy))
	}
	response.Success(c, http.StatusOK, items, gin.H{
		"page":      normalizedPage,
		"page_size": normalizedPageSize,
		"total":     total,
	})
}

func (h *CommunityHandler) CreatePost(c *gin.Context) {
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid post payload"))
		return
	}

	post, err := h.community.CreatePost(c.Request.Context(), middleware.MustUserID(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.ToPostResponse(post), nil)
}

func (h *CommunityHandler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid post id"))
		return
	}
	post, err := h.community.GetPost(c.Request.Context(), postID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToPostResponse(post), nil)
}

func (h *CommunityHandler) UpdatePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid post id"))
		return
	}
	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid post payload"))
		return
	}

	post, err := h.community.UpdatePost(c.Request.Context(), middleware.MustUserID(c), postID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToPostResponse(post), nil)
}

func (h *CommunityHandler) DeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid post id"))
		return
	}
	if err := h.community.DeletePost(c.Request.Context(), middleware.MustUserID(c), postID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func (h *CommunityHandler) ToggleLike(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid post id"))
		return
	}
	liked, likeCount, err := h.community.ToggleLike(c.Request.Context(), middleware.MustUserID(c), postID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToggleLikeResponse{
		Liked:     liked,
		LikeCount: likeCount,
	}, nil)
}

func (h *CommunityHandler) ListComments(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid post id"))
		return
	}
	comments, err := h.community.ListComments(c.Request.Context(), postID)
	if err != nil {
		response.Error(c, err)
		return
	}
	items := make([]dto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		commentCopy := comment
		items = append(items, dto.ToCommentResponse(&commentCopy))
	}
	response.Success(c, http.StatusOK, items, nil)
}

func (h *CommunityHandler) CreateComment(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid post id"))
		return
	}
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid comment payload"))
		return
	}

	comment, err := h.community.CreateComment(c.Request.Context(), middleware.MustUserID(c), postID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.ToCommentResponse(comment), nil)
}

func (h *CommunityHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid comment id"))
		return
	}
	if err := h.community.DeleteComment(c.Request.Context(), middleware.MustUserID(c), commentID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func (h *CommunityHandler) AskAI(c *gin.Context) {
	var req dto.AskCommunityAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid ai payload"))
		return
	}
	response.Success(c, http.StatusOK, dto.CommunityAIResponse{
		Question:  req.Question,
		Answer:    h.community.AskAI(c.Request.Context(), req.Question),
		CreatedAt: time.Now(),
	}, nil)
}
