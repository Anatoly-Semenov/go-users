package http

import (
	"encoding/json"
	"github.com/anatoly_dev/go-users/app/internal/app"
	"github.com/anatoly_dev/go-users/app/internal/domain/auth"
	"github.com/anatoly_dev/go-users/app/internal/domain/user"
	"github.com/anatoly_dev/go-users/app/pkg/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type UserHandler struct {
	userService   *app.UserService
	metricsHelper *metrics.MetricsHelper
}

func NewUserHandler(userService *app.UserService, metricsHelper *metrics.MetricsHelper) *UserHandler {
	return &UserHandler{
		userService:   userService,
		metricsHelper: metricsHelper,
	}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/users", h.handleUsers)
	mux.HandleFunc("/api/users/", h.handleUserByID)
	mux.HandleFunc("/api/login", h.handleLogin)
	mux.HandleFunc("/api/register", h.handleRegister)

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", h.handleHealthCheck)
}

type registerRequest struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      user.Role `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User  userResponse `json:"user"`
	Token string       `json:"token"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      user.Role `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type updateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metricsHelper.HTTPRegistrationDuration().Observe(time.Since(start).Seconds())
	}()

	if r.Method != http.MethodPost {
		h.metricsHelper.ErrorsTotal().WithLabelValues("http", "method_not_allowed").Inc()
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.metricsHelper.RecordValidationError("request_body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Email == "" {
		h.metricsHelper.RecordValidationError("email")
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if req.Password == "" {
		h.metricsHelper.RecordValidationError("password")
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	params := app.RegisterUserParams{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	createdUser, err := h.userService.Register(r.Context(), params)
	if err != nil {
		h.metricsHelper.RecordUserOperation("register", "failed")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.metricsHelper.RecordUserOperation("register", "success")
	h.metricsHelper.UsersTotal().Inc()
	h.metricsHelper.UsersByRole().WithLabelValues(string(createdUser.Role)).Inc()

	respondWithJSON(w, http.StatusCreated, mapUserToResponse(createdUser))
}

func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metricsHelper.HTTPLoginDuration().Observe(time.Since(start).Seconds())
	}()

	if r.Method != http.MethodPost {
		h.metricsHelper.ErrorsTotal().WithLabelValues("http", "method_not_allowed").Inc()
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.metricsHelper.RecordValidationError("request_body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Email == "" {
		h.metricsHelper.RecordValidationError("email")
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if req.Password == "" {
		h.metricsHelper.RecordValidationError("password")
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	clientIP := getClientIP(r)

	user, token, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.metricsHelper.RecordLoginAttempt(false, clientIP)
		h.metricsHelper.RecordUserOperation("login", "failed")

		switch err {
		case auth.ErrInvalidCredentials:
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		case auth.ErrIPBlocked:
			h.metricsHelper.RecordBlockedRequest("ip_blocked")
			respondWithError(w, http.StatusForbidden, "Your IP address is blocked")
		case auth.ErrTooManyAttempts:
			h.metricsHelper.RecordBruteforceAttempt(clientIP)
			respondWithError(w, http.StatusTooManyRequests, "Too many login attempts. Please try again later.")
		default:
			h.metricsHelper.ErrorsTotal().WithLabelValues("auth", "unknown").Inc()
			respondWithError(w, http.StatusInternalServerError, "Failed to authenticate")
		}
		return
	}

	h.metricsHelper.RecordLoginAttempt(true, clientIP)
	h.metricsHelper.RecordUserOperation("login", "success")

	respondWithJSON(w, http.StatusOK, authResponse{
		User:  mapUserToResponse(user),
		Token: token,
	})
}

func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metricsHelper.HTTPUserLookupDuration().Observe(time.Since(start).Seconds())
	}()

	if r.Method != http.MethodGet {
		h.metricsHelper.ErrorsTotal().WithLabelValues("http", "method_not_allowed").Inc()
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	users, err := h.userService.List(r.Context(), offset, limit)
	if err != nil {
		h.metricsHelper.RecordUserOperation("list", "failed")
		h.metricsHelper.ErrorsTotal().WithLabelValues("http", "internal_error").Inc()
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	h.metricsHelper.RecordUserOperation("list", "success")

	responseUsers := make([]userResponse, 0, len(users))
	for _, u := range users {
		responseUsers = append(responseUsers, mapUserToResponse(u))
	}

	respondWithJSON(w, http.StatusOK, responseUsers)
}

func (h *UserHandler) handleUserByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metricsHelper.HTTPUserLookupDuration().Observe(time.Since(start).Seconds())
	}()

	idStr := r.URL.Path[len("/api/users/"):]
	if idStr == "" {
		h.metricsHelper.RecordValidationError("user_id")
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.metricsHelper.RecordValidationError("user_id")
		respondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetUser(w, r, id)
	case http.MethodPut:
		h.handleUpdateUser(w, r, id)
	case http.MethodDelete:
		h.handleDeleteUser(w, r, id)
	default:
		h.metricsHelper.ErrorsTotal().WithLabelValues("http", "method_not_allowed").Inc()
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) handleGetUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		h.metricsHelper.RecordUserOperation("get", "failed")
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	h.metricsHelper.RecordUserOperation("get", "success")
	respondWithJSON(w, http.StatusOK, mapUserToResponse(user))
}

func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.metricsHelper.RecordValidationError("request_body")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	params := app.UpdateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updatedUser, err := h.userService.Update(r.Context(), id, params)
	if err != nil {
		h.metricsHelper.RecordUserOperation("update", "failed")
		respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	h.metricsHelper.RecordUserOperation("update", "success")
	respondWithJSON(w, http.StatusOK, mapUserToResponse(updatedUser))
}

func (h *UserHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.userService.Delete(r.Context(), id); err != nil {
		h.metricsHelper.RecordUserOperation("delete", "failed")
		respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	h.metricsHelper.RecordUserOperation("delete", "success")
	h.metricsHelper.UsersTotal().Dec()
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metricsHelper.AppHealthCheckDuration().Observe(time.Since(start).Seconds())
	}()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.metricsHelper.UpdateUptime()

	uptime := time.Since(h.metricsHelper.GetAppStartTime())

	response := healthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
	}

	respondWithJSON(w, http.StatusOK, response)
}

func mapUserToResponse(user *user.User) userResponse {
	return userResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, errorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func getClientIP(r *http.Request) string {
	if ip, ok := r.Context().Value("client_ip").(string); ok && ip != "" {
		return ip
	}

	return extractIP(r)
}
