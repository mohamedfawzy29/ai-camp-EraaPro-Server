package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eraasoft/ai-camp-server/internal/database"
	"github.com/eraasoft/ai-camp-server/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AdminHandler handles admin-specific operations like exporting leads and stats.
type AdminHandler struct {
	db *database.Client
}

// NewAdminHandler creates a new AdminHandler instance.
func NewAdminHandler(db *database.Client) *AdminHandler {
	return &AdminHandler{db: db}
}

// ExportCSV streams all registrations formatted as a UTF-8 CSV with BOM for Excel compatibility.
// Route: GET /api/admin/export
func (h *AdminHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	coll := h.db.Collection("registrations")
	findOpts := options.Find().SetSort(bson.D{{Key: "submittedAt", Value: -1}})

	cursor, err := coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		log.Printf("[ERROR] Failed to query registrations for export: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "تعذر جلب البيانات من قاعدة البيانات",
		})
		return
	}
	defer cursor.Close(ctx)

	// Set headers for file download
	filename := fmt.Sprintf("eraasoft_ai_camp_leads_%s.csv", time.Now().Format("2006-01-02_1504"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Write UTF-8 BOM so Microsoft Excel renders Arabic correctly without garbled text
	_, _ = w.Write([]byte("\xef\xbb\xbf"))

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Header row
	header := []string{
		"م",
		"الاسم بالكامل",
		"البريد الإلكتروني",
		"رقم الموبايل",
		"الحالة الدراسية",
		"المؤهل الدراسي",
		"التخصص / Track",
		"مجالات استخدام الـ AI",
		"المستوى",
		"تاريخ ووقت التسجيل (UTC)",
		"عنوان IP",
	}
	if err := csvWriter.Write(header); err != nil {
		log.Printf("[ERROR] Failed to write CSV header: %v", err)
		return
	}

	index := 1
	for cursor.Next(ctx) {
		var reg models.Registration
		if err := cursor.Decode(&reg); err != nil {
			log.Printf("[WARN] Failed to decode registration record: %v", err)
			continue
		}

		statusLabel := translateStatus(reg.Status)
		levelLabel := translateLevel(reg.Level)

		row := []string{
			fmt.Sprintf("%d", index),
			reg.Name,
			reg.Email,
			reg.Phone,
			statusLabel,
			reg.Qualification,
			reg.Track,
			reg.UseCases,
			levelLabel,
			reg.SubmittedAt.Format("2006-01-02 15:04:05"),
			reg.ClientIP,
		}

		if err := csvWriter.Write(row); err != nil {
			log.Printf("[ERROR] Failed to write CSV row: %v", err)
			return
		}
		index++
	}

	if err := cursor.Err(); err != nil {
		log.Printf("[ERROR] Cursor error during export: %v", err)
	}
}

// GetStats returns summary statistics for the registrations.
// Route: GET /api/admin/stats
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	coll := h.db.Collection("registrations")

	totalCount, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "تعذر حساب الإحصائيات",
		})
		return
	}

	// Aggregation for status distribution
	statusCounts := aggregateField(ctx, coll, "status")
	levelCounts := aggregateField(ctx, coll, "level")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"totalRegistrations": totalCount,
			"byStatus":           statusCounts,
			"byLevel":            levelCounts,
			"serverTime":         time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// ListRegistrations returns all registrations as JSON.
// Route: GET /api/admin/list
func (h *AdminHandler) ListRegistrations(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	coll := h.db.Collection("registrations")
	findOpts := options.Find().SetSort(bson.D{{Key: "submittedAt", Value: -1}}).SetLimit(500)

	cursor, err := coll.Find(ctx, bson.M{}, findOpts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "تعذر جلب البيانات",
		})
		return
	}
	defer cursor.Close(ctx)

	var list []models.Registration
	if err := cursor.All(ctx, &list); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "تعذر فك ترميز البيانات",
		})
		return
	}

	if list == nil {
		list = []models.Registration{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(list),
		"data":    list,
	})
}

func aggregateField(ctx context.Context, coll *mongo.Collection, field string) map[string]int64 {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$" + field},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	result := make(map[string]int64)
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("[WARN] Aggregation error for field %s: %v", field, err)
		return result
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&item); err == nil && item.ID != "" {
			result[item.ID] = item.Count
		}
	}

	return result
}

func translateStatus(val string) string {
	switch val {
	case "student":
		return "طالب"
	case "graduate":
		return "خريج"
	case "employee":
		return "موظف"
	default:
		return val
	}
}

func translateLevel(val string) string {
	switch val {
	case "beginner":
		return "مبتدئ"
	case "intermediate":
		return "متوسط"
	case "advanced":
		return "متقدم"
	default:
		return val
	}
}
