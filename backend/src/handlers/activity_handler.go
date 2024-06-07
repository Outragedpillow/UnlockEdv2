package handlers

import (
	"UnlockEdv2/src/models"
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (srv *Server) registerActivityRoutes() {
	srv.Mux.Handle("GET /api/users/{id}/activity", srv.applyMiddleware(http.HandlerFunc(srv.GetActivityByUserID)))
	srv.Mux.Handle("GET /api/users/{id}/daily-activity", srv.applyMiddleware(http.HandlerFunc(srv.GetDailyActivityByUserID)))
	srv.Mux.Handle("GET /api/programs/{id}/activity", srv.applyAdminMiddleware(http.HandlerFunc(srv.HandleGetProgramActivity)))
	srv.Mux.Handle("POST /api/users/{id}/activity", srv.applyAdminMiddleware(http.HandlerFunc(srv.HandleCreateActivity)))
}

/****
 * @Query Params:
 * ?program=: id
 * ?year=: year (default last year)
 ****/
func (srv *Server) GetActivityByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	count, activities, err := srv.Db.GetActivityByUserID(1, 365, userID)
	if err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to get activities")
		return
	}
	if err = srv.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"count":      count,
		"activities": activities,
	}); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to write response")
		log.Error("Failed to write response", err)
	}
}

/****
 * @Query Params:
 * ?year=: year (default last year)
 ****/
 func (srv *Server) GetDailyActivityByUserID(w http.ResponseWriter, r *http.Request) {
	// Parse userID from path
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get year from query param
	yearStr := r.URL.Query().Get("year")

	// Convert year parameter to integer
	var year int
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			srv.ErrorResponse(w, http.StatusBadRequest, "Invalid year parameter")
			return
		}
	}

	// Retrieve daily activities for the given userID and year
	activities, err := srv.Db.GetDailyActivityByUserID(userID, year)
	if err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to get activities")
		return
	}

	// Write response
	if err = srv.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
	}); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to write response")
		log.Error("Failed to write response", err)
	}
}


func (srv *Server) HandleGetProgramActivity(w http.ResponseWriter, r *http.Request) {
	programID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, "Invalid program ID")
		return
	}
	page, perPage := srv.GetPaginationInfo(r)
	count, activities, err := srv.Db.GetActivityByProgramID(page, perPage, programID)
	if err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to get activities")
		return
	}
	if err = srv.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"count":      count,
		"activities": activities,
	}); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to write response")
		log.Error("Failed to write response", err)
	}
}

func (srv *Server) HandleCreateActivity(w http.ResponseWriter, r *http.Request) {
	activity := &models.Activity{}
	if err := json.NewDecoder(r.Body).Decode(activity); err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := srv.Db.CreateActivity(activity); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to create activity")
		return
	}
	if err := srv.WriteResponse(w, http.StatusOK, activity); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, "Failed to write response")
		log.Error("Failed to write response", err)
	}
}
