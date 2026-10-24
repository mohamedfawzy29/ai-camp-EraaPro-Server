# 🚀 EraaSoft AI Camp – Go + MongoDB Backend

سيرفر باك إند فائق السرعة مبني بلغة **Go** وقاعدة بيانات **MongoDB** لإدارة استمارات التسجيل في معسكر الذكاء الاصطناعي (AI Bootcamp).

---

## 📋 المميزات
- ⚡ **أداء عالي واستهلاك خفيف للذاكرة**: مبني بـ Go الخفيفة (~20MB RAM فقط).
- 🛡️ **فحص أمني متكامل للبيانات**: تحقق دقيق من صحة الاسم، الإيميل، ورقم الموبايل المصري من جهة الخادم.
- 🚫 **حماية من التكرار**: منع تسجيل نفس البريد الإلكتروني أكثر من مرة عبر Unique Index.
- 📊 **تصدير Excel / CSV مع دعم كامل للعربية**: ترميز `UTF-8 BOM` لفتح الملف مباشرة داخل Microsoft Excel بدون أي تشويه في الحروف.
- 🔒 **حماية مسارات الإدارة**: حماية مسارات الإحصائيات والتصدير بكلمة سر سرية (`ADMIN_SECRET`).
- 🌐 **دعم كامل لـ CORS**: جاهز للربط الفوري مع الفرونت إند في بيئة التطوير والإنتاج.
- 🐳 **جاهز للنشر السحابي المجاني**: مرفق `Dockerfile` متعدد المراحل لإنتاج حاوية بحجم 15MB فقط جاهزة للنشر على Render أو Koyeb.

---

## 🗂️ هيكل المشروع

```text
server/
├── cmd/
│   └── api/
│       └── main.go                 # نقطة الدخول، إعداد المسارات، الإيقاف الآمن
├── internal/
│   ├── config/
│   │   └── config.go               # تحميل وفحص متغيرات البيئة
│   ├── database/
│   │   └── mongodb.go              # الاتصال بـ MongoDB وإدارة الفهارس
│   ├── models/
│   │   └── registration.go         # نماذج البيانات والتحقق من صحتها
│   ├── handlers/
│   │   ├── registration_handler.go # مسار استقبال وحفظ التسجيلات (POST /api/register)
│   │   └── admin_handler.go        # مسار تصدير CSV ومسار الإحصائيات (GET /api/admin/*)
│   └── middleware/
│       ├── cors.go                 # دعم طلبات Cross-Origin
│       └── auth.go                 # فحص مفتاح الإدارة ADMIN_SECRET
├── .env.example                    # نموذج المتغيرات
├── .env                            # المتغيرات المحلية
├── Dockerfile                      # حاوية الإنتاج الخفيفة
└── go.mod                          # حزم واعتماديات المشروع
```

---

## ⚙️ متغيرات البيئة (`.env`)

| المتغير | الوصف | مثال محلي | مثال إنتاج |
| :--- | :--- | :--- | :--- |
| `PORT` | منفذ السيرفر | `8080` | `8080` |
| `MONGODB_URI` | رابط الاتصال بـ MongoDB | `mongodb://localhost:27017` | `mongodb+srv://user:pass@cluster.mongodb.net/` |
| `DB_NAME` | اسم قاعدة البيانات | `ai_camp` | `ai_camp` |
| `ADMIN_SECRET` | كلمة السر لحماية التصدير والإحصائيات | `change_me_to_a_strong_secret_key` | `سلسلة سرية قوية` |
| `ALLOWED_ORIGINS` | الروابط المسموح لها بالاتصال | `http://localhost:5173` | `https://your-frontend.vercel.app` |

---

## 🔌 نقاط النهاية (API Endpoints)

### 1. تسجيل مشترك جديد (العام)
- **المسار**: `POST /api/register`
- **نوع المحتوى**: `application/json` أو `text/plain`
- **جسم الطلب (Request Body)**:
```json
{
  "name": "أحمد محمد علي",
  "email": "ahmed@example.com",
  "phone": "01012345678",
  "status": "graduate",
  "qualification": "بكالوريوس حاسبات ومعلومات",
  "track": "Front-End Development",
  "useCases": "البرمجة، زيادة الإنتاجية",
  "level": "intermediate"
}
```
- **الرد في حالة النجاح (201 Created)**:
```json
{
  "success": true,
  "message": "تم تسجيل بياناتك بنجاح في AI Bootcamp!",
  "id": "6719b2e7c41398246f901a2b"
}
```

---

### 2. تصدير ملف Excel / CSV (محمي للـ Admin)
- **المسار**: `GET /api/admin/export?secret=YOUR_ADMIN_SECRET`
- **الوصف**: ينزل ملف CSV يحتوي على جميع المسجلين مرتبين من الأحدث إلى الأقدم، مع دعم كامل للغة العربية في Microsoft Excel.
- **طريقة التحميل السريعة**: يمكنك ببساطة فتح الرابط التالي في المتصفح:
  `http://localhost:8080/api/admin/export?secret=YOUR_ADMIN_SECRET`

---

### 3. إحصائيات سريعة للـ Admin
- **المسار**: `GET /api/admin/stats?secret=YOUR_ADMIN_SECRET`
- **الرد**:
```json
{
  "success": true,
  "data": {
    "totalRegistrations": 142,
    "byStatus": {
      "student": 80,
      "graduate": 50,
      "employee": 12
    },
    "byLevel": {
      "beginner": 45,
      "intermediate": 75,
      "advanced": 22
    },
    "serverTime": "2026-10-24T02:00:00Z"
  }
}
```

---

### 4. فحص صحة السيرفر
- **المسار**: `GET /api/health`
- **الرد**:
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2026-10-24T02:00:00Z"
}
```

---

## 🚀 كيفية التشغيل محلياً

```bash
# 1. الدخول لمجلد السيرفر
cd server

# 2. تشغيل السيرفر مباشرة
go run cmd/api/main.go
```

أو عبر Docker:
```bash
docker build -t ai-camp-server .
docker run -p 8080:8080 --env-file .env ai-camp-server
```

---

## ☁️ خطوات الرفع المجاني على Render.com

1. أنشئ حساباً مجانياً على [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) وأنشئ قاعدة بيانات مجانية (M0 Free Tier)، وانسخ رابط الـ Connection String.
2. ارفع مجلد `server/` إلى مستودع على GitHub.
3. ادخل على [Render.com](https://render.com) واختر **New Web Service**.
4. اربط المستودع، وسيختار Render إعدادات Go تلقائياً:
   - **Build Command**: `go build -o server cmd/api/main.go`
   - **Start Command**: `./server`
5. في قسم **Environment Variables**، أضف:
   - `MONGODB_URI` = رابط MongoDB Atlas الخاص بك
   - `DB_NAME` = `ai_camp`
   - `ADMIN_SECRET` = كلمة سر قوية من اختيارك
   - `ALLOWED_ORIGINS` = رابط موقع الواجهة على Vercel
6. اضغط **Deploy**. سيعطيك Render رابطاً مجانياً مثل:
   `https://ai-camp-server.onrender.com`
7. في مشروع الـ Frontend، ضع هذا الرابط في ملف `.env`:
   ```env
   VITE_SUBMIT_URL=https://ai-camp-server.onrender.com/api/register
   ```
