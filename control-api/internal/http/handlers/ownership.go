package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cp0x-org/xpaywall/control-api/internal/http/middleware"
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

// isSuperadmin reports whether the caller has the superadmin role.
func isSuperadmin(c *gin.Context) bool {
	v, exists := c.Get("role")
	if !exists {
		return false
	}
	role, ok := v.(string)
	return ok && role == middleware.RoleSuperadmin
}

// requireProjectOwner authorizes a mutating action on a project. It is a pure
// ownership check: the caller must equal projects.owner_user_id. Role grants no
// access to other users' project data — superadmin is just an owner of its own.
// On failure responds with 401/403/404/500 and returns false.
func (h *Handler) requireProjectOwner(c *gin.Context, projectID uuid.UUID) bool {
	callerID, ok := callerUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}

	project, err := h.q.GetProject(c.Request.Context(), projectID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return false
	}

	if project.OwnerUserID == nil || *project.OwnerUserID != callerID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only the project owner can perform this action"})
		return false
	}

	return true
}

// requireGlobalEntityMutate authorizes update/delete of a global-capable entity
// (payment method / asset / facilitator). Rules:
//   - delete of a global entity → superadmin only.
//   - otherwise: owner of the entity, or superadmin.
//
// On failure responds with 403 and returns false.
func (h *Handler) requireGlobalEntityMutate(c *gin.Context, isGlobal bool, ownerID *uuid.UUID, isDelete bool) bool {
	if isSuperadmin(c) {
		return true
	}

	callerID, ok := callerUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}

	if isGlobal {
		msg := "only a superadmin can modify a global entity"
		if isDelete {
			msg = "global payments cannot be deleted"
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": msg})
		return false
	}

	if ownerID == nil || *ownerID != callerID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only the owner can perform this action"})
		return false
	}

	return true
}

// resolveIsGlobal decides whether a new global-capable entity is global.
// Only a superadmin may mark an entity global; for everyone else it is forced false.
func resolveIsGlobal(c *gin.Context, requested bool) bool {
	return requested && isSuperadmin(c)
}

// canSeeGlobalEntity reports whether the caller may view a global-capable entity.
func canSeeGlobalEntity(c *gin.Context, isGlobal bool, ownerID *uuid.UUID) bool {
	if isGlobal || isSuperadmin(c) {
		return true
	}
	callerID, ok := callerUserID(c)
	return ok && ownerID != nil && *ownerID == callerID
}
