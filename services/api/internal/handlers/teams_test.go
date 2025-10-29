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

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
)

// ============================================================================
// ListTeams Tests
// ============================================================================

func TestListTeams_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()

	teams := []db.Team{
		{
			ID:          team1ID,
			OrgID:       orgID,
			Name:        "Engineering",
			Description: stringPtr("Software development team"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			UpdatedAt:   pgtype.Timestamp{Valid: true},
		},
		{
			ID:          team2ID,
			OrgID:       orgID,
			Name:        "Product",
			Description: nil,
			CreatedAt:   pgtype.Timestamp{Valid: true},
			UpdatedAt:   pgtype.Timestamp{Valid: true},
		},
	}

	mockDB.EXPECT().
		ListTeams(gomock.Any(), orgID).
		Return(teams, nil)

	// Expect member count queries for each team
	mockDB.EXPECT().
		CountEmployeesByTeam(gomock.Any(), pgtype.UUID{Bytes: team1ID, Valid: true}).
		Return(int64(10), nil)
	mockDB.EXPECT().
		CountEmployeesByTeam(gomock.Any(), pgtype.UUID{Bytes: team2ID, Valid: true}).
		Return(int64(7), nil)

	// Expect agent config count queries for each team
	mockDB.EXPECT().
		CountTeamAgentConfigs(gomock.Any(), team1ID).
		Return(int64(3), nil)
	mockDB.EXPECT().
		CountTeamAgentConfigs(gomock.Any(), team2ID).
		Return(int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListTeams(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListTeamsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Teams, 2)
	assert.Equal(t, "Engineering", response.Teams[0].Name)
	assert.Equal(t, "Product", response.Teams[1].Name)
	assert.NotNil(t, response.Teams[0].Description)
	assert.Equal(t, "Software development team", *response.Teams[0].Description)
	assert.Nil(t, response.Teams[1].Description)
}

func TestListTeams_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()

	mockDB.EXPECT().
		ListTeams(gomock.Any(), orgID).
		Return([]db.Team{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListTeams(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListTeamsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, response.Total)
	assert.Len(t, response.Teams, 0)
}

// ============================================================================
// CreateTeam Tests
// ============================================================================

func TestCreateTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	description := "Software development team"

	createdTeam := db.Team{
		ID:          teamID,
		OrgID:       orgID,
		Name:        "Engineering",
		Description: &description,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		CreateTeam(gomock.Any(), db.CreateTeamParams{
			OrgID:       orgID,
			Name:        "Engineering",
			Description: &description,
		}).
		Return(createdTeam, nil)

	reqBody := api.CreateTeamRequest{
		Name:        "Engineering",
		Description: &description,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateTeam(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.Team
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Engineering", response.Name)
	assert.NotNil(t, response.Description)
	assert.Equal(t, description, *response.Description)
}

func TestCreateTeam_WithoutDescription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()

	createdTeam := db.Team{
		ID:          teamID,
		OrgID:       orgID,
		Name:        "Product",
		Description: nil,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		CreateTeam(gomock.Any(), db.CreateTeamParams{
			OrgID:       orgID,
			Name:        "Product",
			Description: nil,
		}).
		Return(createdTeam, nil)

	reqBody := api.CreateTeamRequest{
		Name: "Product",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateTeam(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateTeam_MissingName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()

	reqBody := api.CreateTeamRequest{
		Name: "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateTeam(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateTeam_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/teams", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateTeam(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// GetTeam Tests
// ============================================================================

func TestGetTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	description := "Engineering team"

	team := db.Team{
		ID:          teamID,
		OrgID:       orgID,
		Name:        "Engineering",
		Description: &description,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}", handler.GetTeam)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Team
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Engineering", response.Name)
	assert.NotNil(t, response.Description)
	assert.Equal(t, description, *response.Description)
}

func TestGetTeam_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(db.Team{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}", handler.GetTeam)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetTeam_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	orgID := uuid.New()

	r := chi.NewRouter()
	r.Get("/teams/{team_id}", handler.GetTeam)

	req := httptest.NewRequest(http.MethodGet, "/teams/invalid-uuid", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// UpdateTeam Tests
// ============================================================================

func TestUpdateTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	teamID := uuid.New()
	orgID := uuid.New()
	newName := "Engineering Updated"
	newDescription := "Updated description"

	updatedTeam := db.Team{
		ID:          teamID,
		OrgID:       orgID,
		Name:        newName,
		Description: &newDescription,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		UpdateTeam(gomock.Any(), db.UpdateTeamParams{
			ID:          teamID,
			Name:        newName,
			Description: &newDescription,
		}).
		Return(updatedTeam, nil)

	reqBody := api.UpdateTeamRequest{
		Name:        &newName,
		Description: &newDescription,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Patch("/teams/{team_id}", handler.UpdateTeam)

	req := httptest.NewRequest(http.MethodPatch, "/teams/"+teamID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Team
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, newName, response.Name)
	assert.NotNil(t, response.Description)
	assert.Equal(t, newDescription, *response.Description)
}

func TestUpdateTeam_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	teamID := uuid.New()
	newName := "Engineering Updated"

	mockDB.EXPECT().
		UpdateTeam(gomock.Any(), gomock.Any()).
		Return(db.Team{}, pgx.ErrNoRows)

	reqBody := api.UpdateTeamRequest{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Patch("/teams/{team_id}", handler.UpdateTeam)

	req := httptest.NewRequest(http.MethodPatch, "/teams/"+teamID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// DeleteTeam Tests
// ============================================================================

func TestDeleteTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	teamID := uuid.New()

	mockDB.EXPECT().
		DeleteTeam(gomock.Any(), teamID).
		Return(nil)

	r := chi.NewRouter()
	r.Delete("/teams/{team_id}", handler.DeleteTeam)

	req := httptest.NewRequest(http.MethodDelete, "/teams/"+teamID.String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteTeam_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamsHandler(mockDB)

	r := chi.NewRouter()
	r.Delete("/teams/{team_id}", handler.DeleteTeam)

	req := httptest.NewRequest(http.MethodDelete, "/teams/invalid-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
