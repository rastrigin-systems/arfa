package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
)

// ============================================================================
// POST /invitations - Create Invitation Tests
// ============================================================================

func TestCreateInvitation_Success(t *testing.T) {
	// TDD: Write this test FIRST to define expected behavior
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	// Test data
	orgID := uuid.New()
	inviterID := uuid.New()
	roleID := uuid.New()
	teamID := uuid.New()
	email := "newuser@example.com"

	// Mock: Check rate limit (should be under 20)
	mockDB.EXPECT().
		CountInvitationsByOrgToday(gomock.Any(), orgID).
		Return(int64(10), nil)

	// Mock: Create invitation
	mockDB.EXPECT().
		CreateInvitation(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateInvitationParams) (db.Invitation, error) {
			// Verify parameters
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, inviterID, params.InviterID)
			assert.Equal(t, email, params.Email)
			assert.Equal(t, roleID, params.RoleID)
			assert.Equal(t, [16]byte(teamID), params.TeamID.Bytes)
			assert.True(t, params.TeamID.Valid)
			assert.NotEmpty(t, params.Token) // Token should be generated
			assert.Len(t, params.Token, 64)  // 256 bits = 64 hex chars

			return db.Invitation{
				ID:         uuid.New(),
				OrgID:      params.OrgID,
				InviterID:  params.InviterID,
				Email:      params.Email,
				RoleID:     params.RoleID,
				TeamID:     params.TeamID,
				Token:      params.Token,
				Status:     "pending",
				ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
				CreatedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	// Request body
	reqBody := api.CreateInvitationRequest{
		Email:  "newuser@example.com",
		RoleId: roleID,
		TeamId: &teamID,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// HTTP request
	req := httptest.NewRequest(http.MethodPost, "/invitations", bytes.NewReader(bodyBytes))
	req = req.WithContext(context.WithValue(req.Context(), "org_id", orgID))
	req = req.WithContext(context.WithValue(req.Context(), "employee_id", inviterID))
	w := httptest.NewRecorder()

	// Execute
	handler.CreateInvitation(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response api.Invitation
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, orgID, response.OrgId)
	assert.Equal(t, string(email), string(response.Email))
	assert.Equal(t, "pending", string(response.Status))
	assert.NotEmpty(t, response.Token)
}

func TestCreateInvitation_RateLimitExceeded(t *testing.T) {
	// TDD: Test rate limiting (20 invites/day)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	orgID := uuid.New()
	inviterID := uuid.New()

	// Mock: Rate limit exceeded
	mockDB.EXPECT().
		CountInvitationsByOrgToday(gomock.Any(), orgID).
		Return(int64(20), nil)

	// Request
	reqBody := api.CreateInvitationRequest{
		Email:  "newuser@example.com",
		RoleId: uuid.New(),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/invitations", bytes.NewReader(bodyBytes))
	req = req.WithContext(context.WithValue(req.Context(), "org_id", orgID))
	req = req.WithContext(context.WithValue(req.Context(), "employee_id", inviterID))
	w := httptest.NewRecorder()

	// Execute
	handler.CreateInvitation(w, req)

	// Assert
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response api.Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Rate limit")
}

// ============================================================================
// GET /invitations - List Invitations Tests
// ============================================================================

func TestListInvitations_Success(t *testing.T) {
	// TDD: Write this test FIRST
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	orgID := uuid.New()

	// Mock: Count total invitations
	mockDB.EXPECT().
		CountInvitations(gomock.Any(), orgID).
		Return(int64(25), nil)

	// Mock: List invitations with pagination
	mockDB.EXPECT().
		ListInvitations(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.ListInvitationsParams) ([]db.Invitation, error) {
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, int32(10), params.Limit)
			assert.Equal(t, int32(0), params.Offset)

			return []db.Invitation{
				{
					ID:        uuid.New(),
					OrgID:     orgID,
					Email:     "user1@example.com",
					Status:    "pending",
					Token:     "token1",
					ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
					CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
					UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
				},
				{
					ID:        uuid.New(),
					OrgID:     orgID,
					Email:     "user2@example.com",
					Status:    "accepted",
					Token:     "token2",
					ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
					CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
					UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
				},
			}, nil
		})

	// HTTP request with query params
	req := httptest.NewRequest(http.MethodGet, "/invitations?limit=10&offset=0", nil)
	req = req.WithContext(context.WithValue(req.Context(), "org_id", orgID))
	w := httptest.NewRecorder()

	// Execute
	handler.ListInvitations(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response api.InvitationListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Invitations, 2)
	assert.Equal(t, int(25), response.Total)
	assert.Equal(t, int(10), response.Limit)
	assert.Equal(t, int(0), response.Offset)
}

// ============================================================================
// GET /invitations/{token} - Validate Token Tests
// ============================================================================

func TestGetInvitationByToken_Success(t *testing.T) {
	// TDD: Test public token validation endpoint
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	token := "valid-token-123"
	orgID := uuid.New()
	roleID := uuid.New()

	// Mock: Get invitation by token
	mockDB.EXPECT().
		GetInvitationByToken(gomock.Any(), token).
		Return(db.GetInvitationByTokenRow{
			ID:        uuid.New(),
			OrgID:     orgID,
			Email:     "invited@example.com",
			RoleID:    roleID,
			Token:     token,
			Status:    "pending",
			ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
			OrgName:   "Test Org",
			RoleName:  "Member",
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	// HTTP request
	req := httptest.NewRequest(http.MethodGet, "/invitations/"+token, nil)
	w := httptest.NewRecorder()

	// We need to set path param - in real Chi router this is done automatically
	// For testing, we'll pass token via context
	req = req.WithContext(context.WithValue(req.Context(), "token", token))

	// Execute
	handler.GetInvitationByToken(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response api.InvitationDetails
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "invited@example.com", string(response.Email))
	assert.Equal(t, "Test Org", response.OrgName)
	assert.Equal(t, "Member", response.RoleName)
	assert.Equal(t, "pending", string(response.Status))
}

func TestGetInvitationByToken_Expired(t *testing.T) {
	// TDD: Test expired token rejection
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	token := "expired-token"

	// Mock: Get invitation (expired)
	mockDB.EXPECT().
		GetInvitationByToken(gomock.Any(), token).
		Return(db.GetInvitationByTokenRow{
			ID:        uuid.New(),
			Status:    "pending",
			ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(-1 * time.Hour), Valid: true}, // Expired 1 hour ago
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/invitations/"+token, nil)
	req = req.WithContext(context.WithValue(req.Context(), "token", token))
	w := httptest.NewRecorder()

	// Execute
	handler.GetInvitationByToken(w, req)

	// Assert
	assert.Equal(t, http.StatusGone, w.Code)
}

// ============================================================================
// POST /invitations/{token}/accept - Accept Invitation Tests
// ============================================================================

func TestAcceptInvitation_Success(t *testing.T) {
	// TDD: Test invitation acceptance with transaction
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	token := "valid-token"
	orgID := uuid.New()
	roleID := uuid.New()
	invitationID := uuid.New()

	// Mock: Get invitation by token
	mockDB.EXPECT().
		GetInvitationByToken(gomock.Any(), token).
		Return(db.GetInvitationByTokenRow{
			ID:        invitationID,
			OrgID:     orgID,
			Email:     "invited@example.com",
			RoleID:    roleID,
			Token:     token,
			Status:    "pending",
			ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
			OrgName:   "Test Org",
			RoleName:  "Member",
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	// Mock: Create employee
	newEmployeeID := uuid.New()
	mockDB.EXPECT().
		CreateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateEmployeeParams) (db.Employee, error) {
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, "invited@example.com", params.Email)
			assert.Equal(t, roleID, params.RoleID)
			assert.NotEmpty(t, params.PasswordHash)

			return db.Employee{
				ID:           newEmployeeID,
				OrgID:        params.OrgID,
				Email:        params.Email,
				FullName:     params.FullName,
				RoleID:       params.RoleID,
				Status:       "active",
				PasswordHash: params.PasswordHash,
				CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	// Mock: Accept invitation
	mockDB.EXPECT().
		AcceptInvitation(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.AcceptInvitationParams) (db.Invitation, error) {
			assert.Equal(t, token, params.Token)
			assert.Equal(t, [16]byte(newEmployeeID), params.AcceptedBy.Bytes)
			assert.True(t, params.AcceptedBy.Valid)

			return db.Invitation{
				ID:         invitationID,
				Status:     "accepted",
				AcceptedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
				AcceptedBy: pgtype.UUID{Bytes: newEmployeeID, Valid: true},
			}, nil
		})

	// Mock: Create session
	mockDB.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		Return(db.Session{
			ID:        uuid.New(),
			ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
		}, nil)

	// Request body
	reqBody := api.AcceptInvitationRequest{
		FullName: "John Doe",
		Password: "securePassword123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/invitations/"+token+"/accept", bytes.NewReader(bodyBytes))
	req = req.WithContext(context.WithValue(req.Context(), "token", token))
	w := httptest.NewRecorder()

	// Execute
	handler.AcceptInvitation(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response api.AcceptInvitationResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotNil(t, response.Employee)
	assert.NotEmpty(t, response.Token)
}

// ============================================================================
// DELETE /invitations/{id} - Cancel Invitation Tests
// ============================================================================

func TestCancelInvitation_Success(t *testing.T) {
	// TDD: Test invitation cancellation (admin only)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	invitationID := uuid.New()
	orgID := uuid.New()

	// Mock: Get invitation by ID
	mockDB.EXPECT().
		GetInvitationByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.GetInvitationByIDParams) (db.Invitation, error) {
			assert.Equal(t, invitationID, params.ID)
			assert.Equal(t, orgID, params.OrgID)

			return db.Invitation{
				ID:     invitationID,
				OrgID:  orgID,
				Status: "pending",
			}, nil
		})

	// Mock: Cancel invitation
	mockDB.EXPECT().
		CancelInvitation(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CancelInvitationParams) error {
			assert.Equal(t, invitationID, params.ID)
			assert.Equal(t, orgID, params.OrgID)
			return nil
		})

	req := httptest.NewRequest(http.MethodDelete, "/invitations/"+invitationID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "org_id", orgID))
	req = req.WithContext(context.WithValue(req.Context(), "invitation_id", invitationID))
	w := httptest.NewRecorder()

	// Execute
	handler.CancelInvitation(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCancelInvitation_AlreadyAccepted(t *testing.T) {
	// TDD: Cannot cancel accepted invitation
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := NewInvitationHandler(mockDB)

	invitationID := uuid.New()
	orgID := uuid.New()

	// Mock: Get invitation (already accepted)
	mockDB.EXPECT().
		GetInvitationByID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.GetInvitationByIDParams) (db.Invitation, error) {
			assert.Equal(t, invitationID, params.ID)
			assert.Equal(t, orgID, params.OrgID)

			return db.Invitation{
				ID:     invitationID,
				OrgID:  orgID,
				Status: "accepted",
			}, nil
		})

	req := httptest.NewRequest(http.MethodDelete, "/invitations/"+invitationID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "org_id", orgID))
	req = req.WithContext(context.WithValue(req.Context(), "invitation_id", invitationID))
	w := httptest.NewRecorder()

	// Execute
	handler.CancelInvitation(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	var response api.Error
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "already accepted")
}
