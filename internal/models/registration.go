package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^(\+?20|0020)?0?1[0125]\d{8}$`)
)

// Registration represents a single submission document stored in MongoDB.
type Registration struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email" json:"email"`
	Phone         string             `bson:"phone" json:"phone"`
	Status        string             `bson:"status" json:"status"`
	Qualification string             `bson:"qualification" json:"qualification"`
	Track         string             `bson:"track" json:"track"`
	UseCases      string             `bson:"useCases" json:"useCases"`
	Level         string             `bson:"level" json:"level"`
	SubmittedAt   time.Time          `bson:"submittedAt" json:"submittedAt"`
	ClientIP      string             `bson:"clientIP,omitempty" json:"clientIP,omitempty"`
	UserAgent     string             `bson:"userAgent,omitempty" json:"userAgent,omitempty"`
}

// RegistrationInput is the JSON payload received from the frontend form.
type RegistrationInput struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
	Qualification string `json:"qualification"`
	Track         string `json:"track"`
	UseCases      string `json:"useCases"`
	Level         string `json:"level"`
	SubmittedAt   string `json:"submittedAt"`
}

// Validate checks that required fields are present and correctly formatted.
func (r *RegistrationInput) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if len([]rune(r.Name)) < 3 {
		return errors.New("من فضلك اكتب اسمك بالكامل (3 أحرف على الأقل)")
	}

	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	if !emailRegex.MatchString(r.Email) {
		return errors.New("البريد الإلكتروني غير صحيح")
	}

	cleanPhone := strings.ReplaceAll(r.Phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	if !phoneRegex.MatchString(cleanPhone) {
		return errors.New("رقم الموبايل غير صحيح (يجب أن يكون رقم مصري مكون من 11 رقم)")
	}
	r.Phone = cleanPhone

	r.Status = strings.TrimSpace(r.Status)
	if r.Status == "" {
		return errors.New("من فضلك اختر حالتك الدراسية")
	}

	r.Qualification = strings.TrimSpace(r.Qualification)
	if r.Qualification == "" {
		return errors.New("من فضلك اكتب مؤهلك الدراسي")
	}

	r.Track = strings.TrimSpace(r.Track)
	if r.Track == "" {
		return errors.New("من فضلك اكتب تخصصك أو الـ Track")
	}

	r.UseCases = strings.TrimSpace(r.UseCases)
	if r.UseCases == "" {
		return errors.New("من فضلك اختر مجالات استخدام الـ AI")
	}

	r.Level = strings.TrimSpace(r.Level)
	if r.Level == "" {
		return errors.New("من فضلك اختر مستواك في استخدام الـ AI")
	}

	return nil
}

// ToRegistration converts the validated input into the database document model.
func (r *RegistrationInput) ToRegistration(clientIP, userAgent string) Registration {
	submittedTime := time.Now().UTC()
	if r.SubmittedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, r.SubmittedAt); err == nil {
			submittedTime = parsed.UTC()
		}
	}

	return Registration{
		Name:          r.Name,
		Email:         r.Email,
		Phone:         r.Phone,
		Status:        r.Status,
		Qualification: r.Qualification,
		Track:         r.Track,
		UseCases:      r.UseCases,
		Level:         r.Level,
		SubmittedAt:   submittedTime,
		ClientIP:      clientIP,
		UserAgent:     userAgent,
	}
}
