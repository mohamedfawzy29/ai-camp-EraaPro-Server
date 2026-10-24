package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/eraasoft/ai-camp-server/internal/database"
	"github.com/eraasoft/ai-camp-server/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegistrationHandler handles registration submissions.
type RegistrationHandler struct {
	db *database.Client
}

// NewRegistrationHandler creates a new RegistrationHandler instance.
func NewRegistrationHandler(db *database.Client) *RegistrationHandler {
	return &RegistrationHandler{db: db}
}

// Register handles POST /api/register
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "طريقة الطلب غير مسموح بها، استخدم POST",
		})
		return
	}

	// Limit body size to 1 MB to prevent denial-of-service
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "حجم البيانات كبير جداً أو تعذر قراءة الطلب",
		})
		return
	}

	var input models.RegistrationInput
	if err := json.Unmarshal(body, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "تنسيق البيانات غير صحيح (JSON غير صالح)",
		})
		return
	}

	// Validate fields
	if err := input.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()
	doc := input.ToRegistration(clientIP, userAgent)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	coll := h.db.Collection("registrations")
	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"error":   "هذا البريد الإلكتروني مسجل لدينا بالفعل!",
			})
			return
		}

		log.Printf("[ERROR] Failed to insert registration: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "حدث خطأ في الخادم أثناء حفظ البيانات، يرجى المحاولة لاحقاً",
		})
		return
	}

	insertedID, _ := result.InsertedID.(primitive.ObjectID)

	log.Printf("[REGISTRATION] New lead registered: %s (%s) - ID: %s", doc.Name, doc.Email, insertedID.Hex())

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "تم تسجيل بياناتك بنجاح في AI Bootcamp!",
		"id":      insertedID.Hex(),
	})
}

// getClientIP extracts the real client IP from reverse proxies or headers.
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// writeJSON is a helper to write standard JSON responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] writeJSON encode failed: %v", err)
	}
}
