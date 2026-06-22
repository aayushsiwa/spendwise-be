package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetSummary(c *gin.Context) {
	from := c.DefaultQuery("from", time.Now().Format("2006-01")+"-01")
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	categoryFilter := c.Query("category")
	typeFilter := c.Query("type")

	summary, err := h.Service.GetSummary(c.Request.Context(), from, to, categoryFilter, typeFilter)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.Info("Summary retrieved successfully",
		"from", from, "to", to,
		"totalIncome", summary.TotalIncome,
		"totalExpense", summary.TotalExpense,
	)

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}
<<<<<<< HEAD

// nextMonth increments a YYYY-MM string by one month.
// (same implementation as before)
func nextMonth(m string) string {
	t, _ := time.Parse("2006-01", m)
	return t.AddDate(0, 1, 0).Format("2006-01")
}
=======
>>>>>>> 0617a2afde94cf5b86ce3dd3494faae90d7b64cd
