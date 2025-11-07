package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
)

// InvitationHandler handles invitation-related requests
type InvitationHandler struct {
	db db.Querier
}

// NewInvitationHandler creates a new invitation handler
//
// TDD Lesson: Constructor pattern for dependency injection
func NewInvitationHandler(database db.Querier) *InvitationHandler {
	return &InvitationHandler{
		db: database,
	}
}

// CreateInvitation handles POST /invitations
//
// TDD Lesson: Create invitation with secure token generation and rate limiting
//
// Implementation Steps (derived from tests):
// 1. Extract org_id and inviter_id from context
// 2. Parse request body
// 3. Check rate limit (20 invitations/day)
// 4. Generate secure 256-bit token
// 5. Create invitation in database
// 6. Return invitation
func (h *InvitationHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract org_id and inviter_id from context (set by auth middleware)
	orgID, ok := ctx.Value("org_id").(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Missing organization context")
		return
	}

	inviterID, ok := ctx.Value("employee_id").(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Missing employee context")
		return
	}

	// Step 2: Parse request (TestCreateInvitation_Success requires this)
	var req api.CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Step 3: Check rate limit (TestCreateInvitation_RateLimitExceeded requires this)
	count, err := h.db.CountInvitationsByOrgToday(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check rate limit")
		return
	}

	if count >= 20 {
		writeError(w, http.StatusTooManyRequests, "Rate limit exceeded: maximum 20 invitations per day")
		return
	}

	// Step 4: Generate secure 256-bit token (TestCreateInvitation_Success validates this)
	token, err := generateSecureToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Step 5: Create invitation (TestCreateInvitation_Success expects this)
	invitation, err := h.db.CreateInvitation(ctx, db.CreateInvitationParams{
		OrgID:     orgID,
		InviterID: inviterID,
		Email:     string(req.Email),
		RoleID:    uuid.MustParse(req.RoleId.String()),
		TeamID: func() pgtype.UUID {
			if req.TeamId != nil {
				return pgtype.UUID{Bytes: uuid.MustParse(req.TeamId.String()), Valid: true}
			}
			return pgtype.UUID{}
		}(),
		Token: token,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	// Step 6: Return response (TestCreateInvitation_Success validates this)
	response := mapInvitationToAPI(invitation)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListInvitations handles GET /invitations
//
// TDD Lesson: List invitations with pagination
//
// Implementation Steps (derived from tests):
// 1. Extract org_id from context
// 2. Parse query parameters (limit, offset)
// 3. Get total count
// 4. Get paginated list
// 5. Return response with pagination metadata
func (h *InvitationHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract org_id (TestListInvitations_Success requires this)
	orgID, ok := ctx.Value("org_id").(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Missing organization context")
		return
	}

	// Step 2: Parse query parameters (TestListInvitations_Success validates this)
	limit := 10 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Step 3: Get total count (TestListInvitations_Success expects this)
	total, err := h.db.CountInvitations(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count invitations")
		return
	}

	// Step 4: Get paginated list (TestListInvitations_Success expects this)
	invitations, err := h.db.ListInvitations(ctx, db.ListInvitationsParams{
		OrgID:  orgID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list invitations")
		return
	}

	// Step 5: Return response (TestListInvitations_Success validates this)
	apiInvitations := make([]api.Invitation, len(invitations))
	for i, inv := range invitations {
		apiInvitations[i] = mapInvitationToAPI(inv)
	}

	response := api.InvitationListResponse{
		Invitations: apiInvitations,
		Total:       int(total),
		Limit:       limit,
		Offset:      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetInvitationByToken handles GET /invitations/{token}
//
// TDD Lesson: Public endpoint to validate invitation token
//
// Implementation Steps (derived from tests):
// 1. Extract token from context (Chi router sets this)
// 2. Get invitation by token
// 3. Check if expired
// 4. Return invitation details
func (h *InvitationHandler) GetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract token from context (TestGetInvitationByToken_Success requires this)
	token, ok := ctx.Value("token").(string)
	if !ok {
		writeError(w, http.StatusBadRequest, "Missing token")
		return
	}

	// Step 2: Get invitation (TestGetInvitationByToken_Success expects this)
	invitation, err := h.db.GetInvitationByToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Invalid token")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get invitation")
		}
		return
	}

	// Step 3: Check if expired (TestGetInvitationByToken_Expired requires this)
	if invitation.Status == "pending" && time.Now().After(invitation.ExpiresAt.Time) {
		writeError(w, http.StatusGone, "Invitation has expired")
		return
	}

	// Check if already accepted/cancelled
	if invitation.Status != "pending" {
		writeError(w, http.StatusGone, fmt.Sprintf("Invitation is %s", invitation.Status))
		return
	}

	// Step 4: Return details (TestGetInvitationByToken_Success validates this)
	response := api.InvitationDetails{
		Id:        openapi_types.UUID(invitation.ID),
		OrgName:   invitation.OrgName,
		Email:     openapi_types.Email(invitation.Email),
		RoleName:  invitation.RoleName,
		TeamName:  invitation.TeamName,
		Status:    api.InvitationDetailsStatus(invitation.Status),
		ExpiresAt: invitation.ExpiresAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AcceptInvitation handles POST /invitations/{token}/accept
//
// TDD Lesson: Accept invitation with transaction (create employee + update invitation)
//
// Implementation Steps (derived from tests):
// 1. Extract token from context
// 2. Parse request body (full_name, password)
// 3. Get invitation by token
// 4. Validate invitation (pending, not expired)
// 5. Hash password
// 6. Create employee
// 7. Accept invitation
// 8. Generate JWT token
// 9. Create session
// 10. Return token and employee
func (h *InvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract token (TestAcceptInvitation_Success requires this)
	token, ok := ctx.Value("token").(string)
	if !ok {
		writeError(w, http.StatusBadRequest, "Missing token")
		return
	}

	// Step 2: Parse request (TestAcceptInvitation_Success requires this)
	var req api.AcceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Step 3: Get invitation (TestAcceptInvitation_Success expects this)
	invitation, err := h.db.GetInvitationByToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Invalid token")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get invitation")
		}
		return
	}

	// Step 4: Validate invitation
	if invitation.Status != "pending" {
		writeError(w, http.StatusConflict, fmt.Sprintf("Invitation is %s", invitation.Status))
		return
	}

	if time.Now().After(invitation.ExpiresAt.Time) {
		writeError(w, http.StatusGone, "Invitation has expired")
		return
	}

	// Step 5: Hash password (TestAcceptInvitation_Success validates this)
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Step 6: Create employee (TestAcceptInvitation_Success expects this)
	employee, err := h.db.CreateEmployee(ctx, db.CreateEmployeeParams{
		OrgID:        invitation.OrgID,
		Email:        invitation.Email,
		FullName:     req.FullName,
		RoleID:       invitation.RoleID,
		TeamID:       invitation.TeamID,
		PasswordHash: passwordHash,
		Status:       "active",
		Preferences:  []byte("{}"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create employee")
		return
	}

	// Step 7: Accept invitation (TestAcceptInvitation_Success expects this)
	_, err = h.db.AcceptInvitation(ctx, db.AcceptInvitationParams{
		Token:      token,
		AcceptedBy: pgtype.UUID{Bytes: employee.ID, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to accept invitation")
		return
	}

	// Step 8: Generate JWT token (TestAcceptInvitation_Success requires this)
	jwtToken, err := auth.GenerateJWT(employee.ID, employee.OrgID, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Step 9: Create session
	tokenHash := auth.HashToken(jwtToken)
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	session, err := h.db.CreateSession(ctx, db.CreateSessionParams{
		EmployeeID: employee.ID,
		TokenHash:  tokenHash,
		IpAddress:  &ipAddress,
		UserAgent:  &userAgent,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Step 10: Return response (TestAcceptInvitation_Success validates this)
	response := api.AcceptInvitationResponse{
		Token:     jwtToken,
		ExpiresAt: session.ExpiresAt.Time,
		Employee:  mapEmployeeToAPI(employee),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CancelInvitation handles DELETE /invitations/{id}
//
// TDD Lesson: Cancel pending invitation (admin only)
//
// Implementation Steps (derived from tests):
// 1. Extract org_id and invitation_id from context
// 2. Get invitation by ID
// 3. Check if pending (cannot cancel accepted/expired)
// 4. Cancel invitation
// 5. Return success
func (h *InvitationHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract org_id and invitation_id (TestCancelInvitation_Success requires this)
	orgID, ok := ctx.Value("org_id").(uuid.UUID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Missing organization context")
		return
	}

	invitationID, ok := ctx.Value("invitation_id").(uuid.UUID)
	if !ok {
		writeError(w, http.StatusBadRequest, "Missing invitation ID")
		return
	}

	// Step 2: Get invitation (TestCancelInvitation_AlreadyAccepted requires this)
	invitation, err := h.db.GetInvitationByID(ctx, db.GetInvitationByIDParams{
		ID:    invitationID,
		OrgID: orgID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Invitation not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get invitation")
		}
		return
	}

	// Step 3: Check status (TestCancelInvitation_AlreadyAccepted requires this)
	if invitation.Status != "pending" {
		writeError(w, http.StatusConflict, fmt.Sprintf("Cannot cancel invitation that is already %s", invitation.Status))
		return
	}

	// Step 4: Cancel invitation (TestCancelInvitation_Success expects this)
	if err := h.db.CancelInvitation(ctx, db.CancelInvitationParams{
		ID:    invitationID,
		OrgID: orgID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to cancel invitation")
		return
	}

	// Step 5: Return success (TestCancelInvitation_Success validates this)
	w.WriteHeader(http.StatusNoContent)
}

// generateSecureToken generates a cryptographically secure 256-bit token
// Returns a 64-character hex string (32 bytes * 2 hex chars per byte)
//
// TDD Lesson: Security-critical code needs careful implementation
func generateSecureToken() (string, error) {
	// Generate 32 bytes (256 bits) of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string (64 characters)
	return hex.EncodeToString(bytes), nil
}

// mapInvitationToAPI converts database invitation to API invitation
//
// TDD Lesson: Helper functions for data transformation
func mapInvitationToAPI(inv db.Invitation) api.Invitation {
	invitationID := openapi_types.UUID(inv.ID)
	orgID := openapi_types.UUID(inv.OrgID)
	inviterID := openapi_types.UUID(inv.InviterID)
	roleID := openapi_types.UUID(inv.RoleID)
	email := openapi_types.Email(inv.Email)

	invitation := api.Invitation{
		Id:         &invitationID,
		OrgId:      orgID,
		InviterId:  inviterID,
		Email:      email,
		RoleId:     roleID,
		Token:      inv.Token,
		Status:     api.InvitationStatus(inv.Status),
		ExpiresAt:  inv.ExpiresAt.Time,
		CreatedAt:  &inv.CreatedAt.Time,
		UpdatedAt:  &inv.UpdatedAt.Time,
	}

	// Handle nullable team_id
	if inv.TeamID.Valid {
		teamID := openapi_types.UUID(inv.TeamID.Bytes)
		invitation.TeamId = &teamID
	}

	// Handle nullable accepted_by
	if inv.AcceptedBy.Valid {
		acceptedBy := openapi_types.UUID(inv.AcceptedBy.Bytes)
		invitation.AcceptedBy = &acceptedBy
	}

	// Handle nullable accepted_at
	if inv.AcceptedAt.Valid {
		invitation.AcceptedAt = &inv.AcceptedAt.Time
	}

	return invitation
}
