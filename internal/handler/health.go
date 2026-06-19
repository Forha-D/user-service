package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthHandler struct {
	MongoClient    *mongo.Client
	StartTime      time.Time // uptime tracking
	AuthServiceURL string    // if we had an external auth service to check
	HTTPClient     *http.Client
}

func NewHealthHandler(mongoClient *mongo.Client, authServiceURL string) *HealthHandler {
	return &HealthHandler{
		MongoClient:    mongoClient,
		StartTime:      time.Now(),
		AuthServiceURL: authServiceURL,
		HTTPClient:     &http.Client{Timeout: 3 * time.Second},
	}
}

//Response shapes

type pingResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	UptimeMs  int64  `json:"uptime_ms"`
}

type serviceStatus struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Message   string `json:"message,omitempty"` // only shown on failure
}

// public — minimal, safe to expose to internet
type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// detailed — for internal/ops use only
type detailedHealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	UptimeMs  int64                    `json:"uptime_ms"`
	Services  map[string]serviceStatus `json:"services"`
}

// Helpers

func (h *HealthHandler) checkMongo(ctx context.Context) serviceStatus {
	start := time.Now()
	err := h.MongoClient.Ping(ctx, nil)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return serviceStatus{
			Status:    "unhealthy",
			LatencyMs: latency,
			Message:   err.Error(),
		}
	}

	// responding but slow = degraded
	status := "healthy"
	if latency > 200 {
		status = "degraded"
	}

	return serviceStatus{
		Status:    status,
		LatencyMs: latency,
	}
}

func (h *HealthHandler) checkAuthService() serviceStatus {
	start := time.Now()

	resp, err := h.HTTPClient.Get(h.AuthServiceURL + "/ping")
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return serviceStatus{
			Status:    "unhealthy",
			LatencyMs: latency,
			Message:   "auth-service unreachable",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return serviceStatus{
			Status:    "unhealthy",
			LatencyMs: latency,
			Message:   "auth-service returned non-200",
		}
	}

	status := "healthy"
	if latency > 300 { // slightly higher — network hop involved
		status = "degraded"
	}
	return serviceStatus{Status: status, LatencyMs: latency}
}

func overallStatus(services map[string]serviceStatus) (string, int) {
	hasUnhealthy := false
	hasDegraded := false

	for _, s := range services {
		switch s.Status {
		case "unhealthy":
			hasUnhealthy = true
		case "degraded":
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return "unhealthy", http.StatusServiceUnavailable // 503
	}
	if hasDegraded {
		return "degraded", http.StatusOK // 200 — stays in load balancer rotation
	}
	return "healthy", http.StatusOK
}

// Handlers

// Ping — is the process alive? lightweight, no DB call
func (h *HealthHandler) Ping(c echo.Context) error {
	return c.JSON(http.StatusOK, pingResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		UptimeMs:  time.Since(h.StartTime).Milliseconds(),
	})
}

// Live — K8s liveness probe, restarts pod if this fails
func (h *HealthHandler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Ready — K8s readiness probe, removes pod from LB if this fails
func (h *HealthHandler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mongo := h.checkMongo(ctx)
	auth := h.checkAuthService()

	if mongo.Status == "unhealthy" || auth.Status == "unhealthy" {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}

// Health — public minimal response, safe to expose externally
func (h *HealthHandler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	services := map[string]serviceStatus{
		"mongodb":      h.checkMongo(ctx),
		"auth-service": h.checkAuthService(),
	}

	status, httpStatus := overallStatus(services)

	return c.JSON(httpStatus, healthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// HealthDetailed — full details, internal/ops only
func (h *HealthHandler) HealthDetailed(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	services := map[string]serviceStatus{
		"mongodb":      h.checkMongo(ctx),
		"auth-service": h.checkAuthService(),
	}

	status, httpStatus := overallStatus(services)

	return c.JSON(httpStatus, detailedHealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		UptimeMs:  time.Since(h.StartTime).Milliseconds(),
		Services:  services,
	})
}
