package database

import (
	"log"
	"suberes_golang/config"
	"suberes_golang/models"
)

func AutoMigrate() {
	err := config.DB.AutoMigrate(
		&models.BankList{},
		&models.BannerList{},
		&models.BaseModel{},
		&models.BeneficiaryTransaction{},
		&models.CategoryService{},
		&models.ComplainImage{},
		&models.Complain{},
		&models.Country{},
		&models.Coverage{},
		&models.District{},
		&models.GuideTable{},
		&models.HelpTable{},
		&models.LayananService{},
		&models.Message{},
		&models.NewsImageList{},
		&models.NewsList{},
		&models.Notification{},
		&models.OrderChatMessage{},
		&models.OrderChat{},
		&models.OrderRejected{},
		&models.OrderRepeat{},
		&models.OrderSelectedMitra{},
		&models.OrderOffer{},
		&models.OrderTransaction{},
		&models.OrderTransactionRepeat{},
		&models.Payment{},
		&models.PaymentMitra{},
		&models.PrivacyPolicy{},
		&models.Region{},
		&models.Reward{},
		&models.ScheduleParticipant{},
		&models.Schedule{},
		&models.ServiceGuarantee{},
		&models.ServicePromo{},
		&models.Service{},
		&models.SubDistrict{},
		&models.SubPaymentTutorial{},
		&models.SubPayment{},
		&models.SubServiceAdded{},
		&models.SubServiceAdditional{},
		&models.SubService{},
		&models.SubToolCredit{},
		&models.SuberesLogs{},
		&models.SyaratKetentuan{},
		&models.TermsCondition{},
		&models.ToolCredit{},
		&models.Tool{},
		&models.Transaction{},
		&models.UserOTP{},
		&models.UserRating{},
		&models.UserTool{},
		&models.User{},
	)

	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Database migrated successfully")
}
