package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// callerUserID extracts the authenticated user ID from the gin context.
// Returns uuid.Nil and ok=false when the user_id is missing or malformed.
func callerUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return id, true
}

// requireProjectOwner authorizes a mutating action on a project.
// Returns true to proceed. On failure, responds with 401/403/404/500 and returns false.
//
// Rules:
//   - Superadmin (caller user_id == uuid.Nil) is always allowed.
//   - Projects with NULL owner_user_id are treated as legacy/unowned and allowed.
//   - Otherwise the caller must equal projects.owner_user_id.
func (h *Handler) requireProjectOwner(c *gin.Context, projectID uuid.UUID) bool {
	callerID, ok := callerUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}

	if callerID == uuid.Nil {
		return true
	}

	project, err := h.q.GetProject(c.Request.Context(), projectID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return false
	}

	if project.OwnerUserID == nil {
		return true
	}

	if *project.OwnerUserID != callerID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only the project owner can perform this action"})
		return false
	}

	return true
}
