package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CMSgov/bcda-app/bcda/database"
	"github.com/CMSgov/bcda-app/bcda/models"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	rr *httptest.ResponseRecorder
	db *gorm.DB
}

func (s *APITestSuite) SetupTest() {
	models.InitializeGormModels()
	s.db = database.GetGORMDbConnection()

	s.rr = httptest.NewRecorder()
}

func (s *APITestSuite) TestBulkRequestMissingToken() {
	req, err := http.NewRequest("GET", "/api/v1/Patient/$export", nil)
	assert.Nil(s.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(bulkRequest)

	handler.ServeHTTP(rr, req)

	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *APITestSuite) TestJobStatusPending() {
	j := models.Job{
		AcoID:      uuid.Parse("DBBD1CE1-AE24-435C-807D-ED45953077D3"),
		UserID:     uuid.Parse("82503A18-BF3B-436D-BA7B-BAE09B7FFD2F"),
		RequestURL: "/api/v1/Patient/$export",
		Status:     "Pending",
	}
	s.db.Save(&j)

	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%d", j.ID), nil)
	assert.Nil(s.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(jobStatus)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobId", fmt.Sprint(j.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(rr, req)

	assert.Equal(s.T(), http.StatusAccepted, rr.Code)
	assert.Equal(s.T(), "Pending", rr.Header().Get("X-Progress"))

	s.db.Delete(&j)
}

func (s *APITestSuite) TestJobStatusInProgress() {
	j := models.Job{
		AcoID:      uuid.Parse("DBBD1CE1-AE24-435C-807D-ED45953077D3"),
		UserID:     uuid.Parse("82503A18-BF3B-436D-BA7B-BAE09B7FFD2F"),
		RequestURL: "/api/v1/Patient/$export",
		Status:     "In Progress",
	}
	s.db.Save(&j)

	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%d", j.ID), nil)
	assert.Nil(s.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(jobStatus)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobId", fmt.Sprint(j.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(rr, req)

	assert.Equal(s.T(), http.StatusAccepted, rr.Code)
	assert.Equal(s.T(), "In Progress", rr.Header().Get("X-Progress"))

	s.db.Delete(&j)
}

func (s *APITestSuite) TestJobStatusFailed() {
	j := models.Job{
		AcoID:      uuid.Parse("DBBD1CE1-AE24-435C-807D-ED45953077D3"),
		UserID:     uuid.Parse("82503A18-BF3B-436D-BA7B-BAE09B7FFD2F"),
		RequestURL: "/api/v1/Patient/$export",
		Status:     "Failed",
	}

	s.db.Save(&j)

	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%d", j.ID), nil)
	assert.Nil(s.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(jobStatus)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobId", fmt.Sprint(j.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(rr, req)

	assert.Equal(s.T(), http.StatusInternalServerError, rr.Code)

	s.db.Delete(&j)
}

func (s *APITestSuite) TestJobStatusCompleted() {
	j := models.Job{
		AcoID:      uuid.Parse("DBBD1CE1-AE24-435C-807D-ED45953077D3"),
		UserID:     uuid.Parse("82503A18-BF3B-436D-BA7B-BAE09B7FFD2F"),
		RequestURL: "/api/v1/Patient/$export",
		Status:     "Completed",
	}
	s.db.Save(&j)

	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%d", j.ID), nil)
	assert.Nil(s.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(jobStatus)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("jobId", fmt.Sprint(j.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ServeHTTP(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Equal(s.T(), "application/json", rr.Header().Get("Content-Type"))

	var rb bulkResponseBody
	err = json.Unmarshal(rr.Body.Bytes(), &rb)
	if err != nil {
		s.Error(err)
	}

	assert.Equal(s.T(), j.RequestURL, rb.RequestURL)
	assert.Equal(s.T(), true, rb.RequiresAccessToken)
	assert.Equal(s.T(), "ExplanationOfBenefit", rb.Files[0].Type)
	assert.Equal(s.T(), "http:///data/DBBD1CE1-AE24-435C-807D-ED45953077D3.ndjson", rb.Files[0].URL)
	assert.Empty(s.T(), rb.Errors)

	s.db.Delete(&j)
}

func (s *APITestSuite) TestServeData() {
	req, err := http.NewRequest("GET", "/api/v1/data/DBBD1CE1-AE24-435C-807D-ED45953077D3.ndjson", nil)
	assert.Nil(s.T(), err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveData)

	handler.ServeHTTP(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), `{"resourceType": "Bundle", "total": 33, "entry": [{"resource": {"status": "active", "diagnosis": [{"diagnosisCodeableConcept": {"coding": [{"system": "http://hl7.org/fhir/sid/icd-9-cm", "code": "2113"}]},`)
}

func (s *APITestSuite) TestGetToken() {}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}