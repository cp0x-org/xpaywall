package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cp0x-org/xpaywall/control-api/internal/http/middleware"
)

// newCtx builds a gin context carrying the given caller identity, mirroring what
// the JWT middleware sets. role="" / id=uuid.Nil model a missing claim.
func newCtx(role string, id uuid.UUID) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if role != "" {
		c.Set("role", role)
	}
	c.Set("user_id", id)
	return c
}

func TestResolveIsGlobal_OnlySuperadminCanMarkGlobal(t *testing.T) {
	user := newCtx("user", uuid.New())
	super := newCtx(middleware.RoleSuperadmin, uuid.New())

	// A regular user requesting is_global must be downgraded to false — this is the
	// rule that prevents users from publishing entities to everyone.
	if resolveIsGlobal(user, true) {
		t.Fatal("regular user must not be able to mark an entity global")
	}
	if !resolveIsGlobal(super, true) {
		t.Fatal("superadmin must be able to mark an entity global")
	}
	if resolveIsGlobal(super, false) {
		t.Fatal("is_global=false request must stay false even for superadmin")
	}
}

func TestCanSeeGlobalEntity(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()

	// Global entities are visible to everyone; personal entities only to their owner
	// (and to any superadmin). This is the visibility contract the list/get endpoints rely on.
	if !canSeeGlobalEntity(newCtx("user", other), true, &owner) {
		t.Fatal("global entity must be visible to a non-owner")
	}
	if !canSeeGlobalEntity(newCtx("user", owner), false, &owner) {
		t.Fatal("personal entity must be visible to its owner")
	}
	if canSeeGlobalEntity(newCtx("user", other), false, &owner) {
		t.Fatal("personal entity must be hidden from a non-owner")
	}
	if !canSeeGlobalEntity(newCtx(middleware.RoleSuperadmin, other), false, &owner) {
		t.Fatal("superadmin must see any personal entity")
	}
}

func TestRequireGlobalEntityMutate_DeleteRules(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()

	cases := []struct {
		name     string
		ctx      *gin.Context
		isGlobal bool
		ownerID  *uuid.UUID
		isDelete bool
		want     bool
	}{
		// Deleting a global entity is superadmin-only — the core of requirement 7.
		{"user deletes global → denied", newCtx("user", owner), true, &owner, true, false},
		{"superadmin deletes global → allowed", newCtx(middleware.RoleSuperadmin, other), true, &owner, true, true},
		{"owner deletes own personal → allowed", newCtx("user", owner), false, &owner, true, true},
		{"non-owner deletes personal → denied", newCtx("user", other), false, &owner, true, false},
		{"owner updates own personal → allowed", newCtx("user", owner), false, &owner, false, true},
		{"user updates global → denied", newCtx("user", owner), true, &owner, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{}
			if got := h.requireGlobalEntityMutate(tc.ctx, tc.isGlobal, tc.ownerID, tc.isDelete); got != tc.want {
				t.Fatalf("requireGlobalEntityMutate = %v, want %v", got, tc.want)
			}
		})
	}
}
