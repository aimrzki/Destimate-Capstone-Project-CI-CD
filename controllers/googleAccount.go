package controllers

// Import paket-paket yang dibutuhkan
import (
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"math/rand"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"time"
)

// GoogleConfig adalah konfigurasi OAuth untuk otentikasi Google
var GoogleConfig = &oauth2.Config{
	ClientID:     "861423210000-uuujnme01jdll7s353s7680f4gjac986.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL:  "https://destimate.uc.r.appspot.com/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// GoogleUser adalah struktur data yang mewakili respons dari Google setelah otentikasi
type GoogleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Endpoint untuk menginisiasi otentikasi Google
func GoogleAuthInitiate(c echo.Context) error {
	url := GoogleConfig.AuthCodeURL("state")
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// Endpoint yang diarahkan oleh Google setelah otentikasi berhasil
func GoogleAuthCallback(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan kode otentikasi dari parameter query
		code := c.QueryParam("code")

		// Menukarkan kode otentikasi dengan token akses dari Google
		token, err := GoogleConfig.Exchange(c.Request().Context(), code)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to exchange code for token"})
		}

		// Mendapatkan informasi pengguna dari Google menggunakan token akses
		googleUser, err := getGoogleUserInfo(token)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to get user info from Google"})
		}

		// Membuat atau mengambil pengguna dari database berdasarkan informasi Google
		user, err := createUserFromGoogle(db, googleUser)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to create user"})
		}

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(user.Username, secretKey)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to generate token"})
		}

		// Mengirim respons sukses bersama dengan token
		return c.JSON(http.StatusOK, map[string]interface{}{"token": tokenString, "id": user.ID, "message": "User signup/login with Google successful"})
	}
}

// Fungsi untuk mendapatkan informasi pengguna dari Google menggunakan token akses
func getGoogleUserInfo(token *oauth2.Token) (*GoogleUser, error) {
	client := GoogleConfig.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&googleUser)
	if err != nil {
		return nil, err
	}

	return &googleUser, nil
}

// Fungsi untuk menghasilkan string acak
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}

func generateRandomPhoneNumber(length int) string {
	const charset = "0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}

// Fungsi untuk membuat atau mengambil pengguna dari database berdasarkan informasi Google
func createUserFromGoogle(db *gorm.DB, googleUser *GoogleUser) (*model.User, error) {
	// Cek apakah pengguna sudah ada dalam database
	var existingUser model.User
	result := db.Where("email = ?", googleUser.Email).First(&existingUser)
	if result.Error == nil {
		// Pengguna sudah ada, kembalikan pengguna yang sudah ada
		return &existingUser, nil
	}

	// Pengguna belum ada, buat pengguna baru
	newUser := model.User{
		Email:             googleUser.Email,
		Name:              googleUser.Name,
		Username:          googleUser.ID,                 // Atur sesuai kebutuhan, misalnya menggunakan ID Google sebagai username
		Password:          generateRandomString(10),      // Atur password sesuai kebutuhan atau gunakan nilai default
		PhoneNumber:       generateRandomPhoneNumber(10), // Atur nomor telepon sesuai kebutuhan atau gunakan nilai default
		IsVerified:        true,                          // Atur nilai default atau sesuai kebutuhan
		VerificationToken: "",                            // Atur sesuai kebutuhan atau gunakan nilai default
	}

	// Simpan pengguna baru ke database
	result = db.Create(&newUser)
	if result.Error != nil {
		return nil, result.Error
	}

	return &newUser, nil
}
