package routes

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"myproject/controllers"
	"net/http"
	"os"
)

func ServeHTML(c echo.Context) error {
	htmlData, err := ioutil.ReadFile("index.html")
	if err != nil {
		return err
	}
	return c.HTML(http.StatusOK, string(htmlData))
}

func getSecretKeyFromEnv() string {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY tidak ditemukan di .env")
	}
	return secretKey
}

func SetupRoutes(e *echo.Echo, db *gorm.DB) {
	e.Use(Logger())
	secretKey := []byte(getSecretKeyFromEnv())

	//Integrate with OAuth Google Account
	e.GET("/auth/google/initiate", controllers.GoogleAuthInitiate)
	e.GET("/auth/google/callback", controllers.GoogleAuthCallback(db, secretKey))

	//All User
	e.POST("/signup", controllers.Signup(db, secretKey))                                                // Register - Mobile
	e.GET("/", ServeHTML)                                                                               // Serve HTML - Email
	e.GET("/verify", controllers.VerifyEmail(db))                                                       // Verify email - Email
	e.GET("/categories", controllers.GetCategories(db, secretKey))                                      // Menampilkan seluruh category yang tersedia
	e.GET("/tourism-attractions", controllers.GetWisatas(db, secretKey))                                // Menampilkan seluruh tempat wisata yang ada - CMS & Mobile
	e.GET("/tourism-attractions/:id", controllers.GetWisataByID(db, secretKey))                         // Menampilkan detail tempat wisata berdasarkan id nya - CMS & Mobile
	e.GET("/carbonfootprints/:wisata_id", controllers.GetTotalCarbonFootprintByWisataID(db, secretKey)) // Menampilkan total carbon footprint pada detail tempat wisata
	e.GET("/promos", controllers.GetPromos(db, secretKey))                                              // Menampilkan seluruh promo yang tersedia - CMS & Mobile
	e.GET("/promos/:id", controllers.GetPromoByID(db, secretKey))                                       // Menampilkan data detail promo yang tersedia - CMS & Mobile

	//Landing Page
	e.POST("/cooperation", controllers.CreateCooperationMessage(db)) // Mengirimkan pesan kepada destimate dari landingpage

	//Dashboard CMS
	e.GET("/dashboard", controllers.GetAdminDashboardData(db, secretKey)) // Menampilkan data grafik dan angka di dashboard yang berisi (Pendapatan, total pengguna, total destinasi, total pengunjung, total pemesanan)
	e.GET("/top/wisata", controllers.GetTopWisata(db, secretKey))         // Menampilkan data tempat wisata yang paling banyak dikunjungi
	e.GET("/top/emition", controllers.GetTopEmition(db, secretKey))       // Menampilkan data user dengan perolehan emisi karbon paling sedikit

	//Admin CMS
	e.POST("/admins/signin", controllers.AdminSignin(db, secretKey))                      // Login Admin - CMS
	e.GET("/profile", controllers.GetProfile(db, secretKey))                              // Get profile admin - CMS
	e.GET("/admin", controllers.GetAllAdminsByAdmin(db, secretKey))                       // Menampilkan seluruh data admin - CMS
	e.GET("/admins/users", controllers.GetAllUsersByAdmin(db, secretKey))                 // Menampilkan seluruh data user - CMS
	e.PUT("/admins/users/:id", controllers.EditUserByAdmin(db, secretKey))                // Mengubah data user - CMS
	e.DELETE("/admins/users/:id", controllers.DeleteUserByAdmin(db, secretKey))           // Menghapus akun user oleh admin - CMS
	e.DELETE("/admins/:id", controllers.DeleteAdminByAdmin(db, secretKey))                // Menghapus akun admin oleh admin - CMS
	e.POST("/categories", controllers.CreateCategoryByAdmin(db, secretKey))               // Membuat category baru oleh user - CMS
	e.PUT("/categories/:id", controllers.UpdateCategoryByAdmin(db, secretKey))            // Mengupdate category yang sudah ada - CMS
	e.DELETE("/categories/:id", controllers.DeleteCategoryByAdmin(db, secretKey))         // Menghapus category yang sudah ada - CMS
	e.POST("/tourism-attractions", controllers.CreateWisata(db, secretKey))               // Menambahkan tempat wisata baru - CMS
	e.PUT("/tourism-attractions/:id", controllers.UpdateWisata(db, secretKey))            // Mengedit data tempat wisata yang sudah ada - CMS
	e.DELETE("/tourism-attractions/:id", controllers.DeleteWisata(db, secretKey))         // Menghapus tempat wisata yang sudah ada - CMS
	e.POST("/promos", controllers.CreatePromo(db, secretKey))                             // Membuat promo baru di platform - CMS
	e.PUT("/promos/:id", controllers.EditPromo(db, secretKey))                            // Mengubah data promo yang sudah ada - CMS
	e.DELETE("/promos/:id", controllers.DeletePromoByAdmin(db, secretKey))                // Menghapus data promo yang sudah ada - CMS
	e.GET("/tickets", controllers.GetAllTicketsByAdmin(db, secretKey))                    // Menampilkan seluruh transaksi yang ada di platform - CMS
	e.GET("/tickets/:invoiceNumber", controllers.GetTicketByInvoiceNumber(db, secretKey)) // Menampilkan detail transaksi yang ada di platform - CMS
	e.PUT("/tickets/:invoiceId", controllers.UpdatePaidStatus(db, secretKey))             // Mengubah status pembayaran - CMS
	e.DELETE("/tickets/:invoice_number", controllers.DeleteTicketByAdmin(db, secretKey))  // Menghapus ticket transaksi oleh admin - CMS
	e.POST("/terms-and-conditions", controllers.CreateTermCondition(db, secretKey))       //  Membuat term and condition baru - CMS
	e.PUT("/terms-and-conditions/:id", controllers.EditTermCondition(db, secretKey))      //  Mengubah term and condition yang sudah ada - CMS
	e.GET("/terms-and-conditions", controllers.GetAllTermCondition(db, secretKey))        //  Melihat seluruh data term and condition yang ada - CMS
	e.GET("/terms-and-conditions/:id", controllers.GetTermConditionByID(db, secretKey))   // Melihat detail  data term and condition - CMS
	e.DELETE("/terms-and-conditions/:id", controllers.DeleteTermCondition(db, secretKey)) // Menghapus term and condition yang ada - CMS
	e.GET("/cooperations", controllers.GetCooperationMessagesByAdmin(db, secretKey))      // Mendapatkan pesan yang user kirim dari landing page

	// Chatbot custom data untuk admin dapat bertanya terkait rekomendasi promo untuk meningkatkan penjualan
	promoChatbotUsecase := controllers.NewPromoChatbotUsecase() // Inisialisasi use case
	e.POST("/users/chatbot", func(c echo.Context) error {
		return controllers.RecommendPromoChatbot(c, promoChatbotUsecase, db, secretKey) // Panggil fungsi dengan use case dan instance Gorm DB
	})

	//User Mobile
	e.POST("/signin", controllers.Signin(db, secretKey))                                                    // Login - Mobile
	e.GET("/users/:user_id", controllers.GetUserDataByID(db, secretKey))                                    // Menampilkan profile user - Mobile
	e.PUT("/users/:id", controllers.EditUser(db, secretKey))                                                // Mengubah profile user - Mobile
	e.PUT("/users/change-password", controllers.ChangePassword(db, secretKey))                              // Mengubah password akun user - Mobile
	e.DELETE("/users/photo/:id", controllers.DeleteUserProfilePhoto(db, secretKey))                         // Menghapus foto profile user - Mobile
	e.PUT("/users/:id/change-location", controllers.EditUserLocation(db, secretKey))                        // Mengubah lokasi user - Mobile
	e.GET("/users/preferences", controllers.GetWisataByCategoryKesukaan(db, secretKey))                     // Menampilkan halaman utama user sesuai wisata preferensinya saat match making diawal - Mobile
	e.GET("/cities", controllers.GetCities(db, secretKey))                                                  // Menampilkan kota yang tersedia dari tempat wisata yang ada - Mobile
	e.POST("/tourism-attractions/booking", controllers.BuyTicket(db, secretKey))                            // Pemesanan tiket oleh user - Mobile
	e.POST("/tourism-attractions/booking/check", controllers.CheckTicketPrice(db, secretKey))               // Melakukan pengecekan harga saat transaksi
	e.DELETE("/tourism-attractions/cancel/:invoice_number", controllers.CancelTicket(db, secretKey))        // Melakukan pembatalan transaksi pada order yang belum dibayar - Mobile
	e.GET("/user/tickets", controllers.GetTicketsByUser(db, secretKey))                                     // Melihat seluruh history pemesanan yang pernah dilakukan user - Mobile
	e.GET("/user/tickets/:invoice_number", controllers.GetTransactionHistoryByInvoiceNumber(db, secretKey)) // Menampilkan detail pemesanan sesuai dengan invoice number - Mobile
	e.GET("/points", controllers.GetUserPoints(db, secretKey))                                              // Menampilkan points yang user miliki dari transaksinya
	e.GET("/points/history", controllers.GetPointsHistory(db, secretKey))                                   // Menampilkan history points yang user miliki
	e.GET("/user/carbonfootprint/:user_id", controllers.GetTotalCarbonFootprintByUser(db, secretKey))       // Menampilkan total karbon footprint yang user hasilkan dari semua perjalanannya
	e.GET("/notifications", controllers.GetUserNotifications(db, secretKey))                                // Menampilkan notifikasi yang user miliki (Notifikasi berhasil bayar & Saat ada promo baru)
	e.PUT("/notifications/:id", controllers.MarkNotificationAsRead(db, secretKey))                          // Menandai notifikasinya sudah dibaca

	// Chatbot untuk user dapat bertanya dengan Debot rekomendasi tempat wisata
	wisataUsecase := controllers.NewWisataUsecase()
	e.POST("/admin/chatbot", func(c echo.Context) error {
		return controllers.RecommendWisata(c, wisataUsecase)
	})

}
