package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hungryhub-team/clickhouse-manager/internal/usecase"
)

type ReportHandler struct {
	reportUsecase     usecase.ReportUsecase
	connectionUsecase *usecase.ConnectionUsecase
}

func NewReportHandler(reportUsecase usecase.ReportUsecase, connectionUsecase *usecase.ConnectionUsecase) *ReportHandler {
	return &ReportHandler{
		reportUsecase:     reportUsecase,
		connectionUsecase: connectionUsecase,
	}
}

func (h *ReportHandler) Register(app *fiber.App) {
	group := app.Group("/connections/:id/reports")
	group.Get("/slow-queries", h.GetSlowQueries)
	group.Get("/resource-stats", h.GetResourceStats)
}

func (h *ReportHandler) GetSlowQueries(c *fiber.Ctx) error {
	connectionID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Connection ID")
	}

	refresh := c.Query("refresh") == "true"
	queryKind := c.Query("queryKind", "all")

	reports, lastRefresh, err := h.reportUsecase.GetTopSlowQueries(c.Context(), connectionID, queryKind, refresh)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if c.Query("format") == "json" || c.Get("Accept") == "application/json" {
		return c.JSON(fiber.Map{
			"data":         reports,
			"last_refresh": lastRefresh,
		})
	}

	// Marshaling reports to JSON for frontend
	reportsJSON, _ := json.Marshal(reports)
	if len(reportsJSON) == 0 {
		reportsJSON = []byte("[]")
	}

	// Fetch connections for sidebar
	connections, _ := h.connectionUsecase.GetAllConnections(c.Context())

	return c.Render("reports/index", fiber.Map{
		"Reports":            string(reportsJSON),
		"ConnectionID":       connectionID,
		"LastRefresh":        lastRefresh,
		"QueryKind":          queryKind,
		"ActiveMenu":         " reports",
		"SidebarConnections": connections,
	}, "layouts/main")
}

// GetResourceStats returns resource usage statistics (memory, queries, merges, fetches)
func (h *ReportHandler) GetResourceStats(c *fiber.Ctx) error {
	connectionID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Connection ID"})
	}

	refresh := c.Query("refresh") == "true"

	// Get process stats from connection usecase
	stats, err := h.connectionUsecase.GetProcessStatsData(c.Context(), connectionID, refresh)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"memory_tracking":     stats.MemoryTracking,
		"queries_in_progress":  stats.QueriesInProgress,
		"background_merges":    stats.BackgroundMerges,
		"background_fetches":   stats.BackgroundFetches,
	})
}
