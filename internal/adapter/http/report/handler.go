package report

import "finance-app-go-demo/internal/application/report"

type Handler struct {
	reportService *report.Service
}

func NewHandler(reportService *report.Service) *Handler {
	return &Handler{
		reportService: reportService,
	}
}
