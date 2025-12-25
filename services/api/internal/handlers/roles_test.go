package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
)

// ============================================================================
// ListRoles Tests
// ============================================================================

func TestListRoles_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	role1ID := uuid.New()
	role2ID := uuid.New()

	roles := []db.Role{
		{
			ID:          role1ID,
			Name:        "Admin",
			Permissions: []byte(`["read","write","delete"]`),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			UpdatedAt:   pgtype.Timestamp{Valid: true},
		},
		{
			ID:          role2ID,
			Name:        "Developer",
			Permissions: []byte(`["read","write"]`),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			UpdatedAt:   pgtype.Timestamp{Valid: true},
		},
	}

	mockDB.EXPECT().
		ListRoles(gomock.Any()).
		Return(roles, nil)

	// Expect employee count queries for each role
	mockDB.EXPECT().
		CountEmployeesByRole(gomock.Any(), role1ID).
		Return(int64(5), nil)
	mockDB.EXPECT().
		CountEmployeesByRole(gomock.Any(), role2ID).
		Return(int64(3), nil)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()

	handler.ListRoles(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Simple inline response struct since api.ListRolesResponse doesn't exist yet
	var response struct {
		Roles []api.Role `json:"roles"`
		Total int        `json:"total"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Roles, 2)
	assert.Equal(t, "Admin", response.Roles[0].Name)
	assert.Equal(t, "Developer", response.Roles[1].Name)
}

func TestListRoles_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	mockDB.EXPECT().
		ListRoles(gomock.Any()).
		Return([]db.Role{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()

	handler.ListRoles(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Roles []api.Role `json:"roles"`
		Total int        `json:"total"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, response.Total)
	assert.Len(t, response.Roles, 0)
}

// ============================================================================
// GetRole Tests
// ============================================================================

func TestGetRole_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	roleID := uuid.New()
	role := db.Role{
		ID:          roleID,
		Name:        "Admin",
		Permissions: []byte(`["read","write","delete"]`),
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetRole(gomock.Any(), roleID).
		Return(role, nil)

	r := chi.NewRouter()
	r.Get("/roles/{role_id}", handler.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Role
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, roleID.String(), response.Id.String())
	assert.Equal(t, "Admin", response.Name)
}

func TestGetRole_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	roleID := uuid.New()

	mockDB.EXPECT().
		GetRole(gomock.Any(), roleID).
		Return(db.Role{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/roles/{role_id}", handler.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetRole_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	r := chi.NewRouter()
	r.Get("/roles/{role_id}", handler.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/roles/invalid-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// CreateRole Tests
// ============================================================================

func TestCreateRole_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	roleID := uuid.New()
	permissions := []string{"read", "write"}

	reqBody := handlers.CreateRoleRequest{
		Name:        "Developer",
		Permissions: permissions,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockDB.EXPECT().
		CreateRole(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx any, params db.CreateRoleParams) (db.Role, error) {
			assert.Equal(t, "Developer", params.Name)
			return db.Role{
				ID:          roleID,
				Name:        params.Name,
				Permissions: params.Permissions,
				CreatedAt:   pgtype.Timestamp{Valid: true},
				UpdatedAt:   pgtype.Timestamp{Valid: true},
			}, nil
		})

	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateRole(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.Role
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Developer", response.Name)
}

func TestCreateRole_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateRole(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateRole_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		reqBody handlers.CreateRoleRequest
	}{
		{
			name: "missing name",
			reqBody: handlers.CreateRoleRequest{
				Permissions: []string{"read"},
			},
		},
		{
			name: "empty name",
			reqBody: handlers.CreateRoleRequest{
				Name:        "",
				Permissions: []string{"read"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			handler := handlers.NewRolesHandler(mockDB)

			bodyBytes, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(string(bodyBytes)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.CreateRole(rec, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		})
	}
}

// ============================================================================
// UpdateRole Tests
// ============================================================================

func TestUpdateRole_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	roleID := uuid.New()
	newName := "Senior Developer"
	newPerms := []string{"read", "write", "execute"}

	reqBody := handlers.UpdateRoleRequest{
		Name:        &newName,
		Permissions: &newPerms,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockDB.EXPECT().
		UpdateRole(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx any, params db.UpdateRoleParams) (db.Role, error) {
			assert.Equal(t, roleID, params.ID)
			assert.Equal(t, newName, params.Name)
			return db.Role{
				ID:          roleID,
				Name:        params.Name,
				Permissions: params.Permissions,
				CreatedAt:   pgtype.Timestamp{Valid: true},
				UpdatedAt:   pgtype.Timestamp{Valid: true},
			}, nil
		})

	r := chi.NewRouter()
	r.Patch("/roles/{role_id}", handler.UpdateRole)

	req := httptest.NewRequest(http.MethodPatch, "/roles/"+roleID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Role
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, newName, response.Name)
}

func TestUpdateRole_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	roleID := uuid.New()
	newName := "Updated Role"

	reqBody := handlers.UpdateRoleRequest{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockDB.EXPECT().
		UpdateRole(gomock.Any(), gomock.Any()).
		Return(db.Role{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Patch("/roles/{role_id}", handler.UpdateRole)

	req := httptest.NewRequest(http.MethodPatch, "/roles/"+roleID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// DeleteRole Tests
// ============================================================================

func TestDeleteRole_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewRolesHandler(mockDB)

	roleID := uuid.New()

	mockDB.EXPECT().
		DeleteRole(gomock.Any(), roleID).
		Return(nil)

	r := chi.NewRouter()
	r.Delete("/roles/{role_id}", handler.DeleteRole)

	req := httptest.NewRequest(http.MethodDelete, "/roles/"+roleID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
