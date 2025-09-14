package services

import (
	"log"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type NotificationService interface {
	CreateNotification(userID uuid.UUID, title, message, link, notifType string)
}

type notificationService struct {
	notifRepo    repositories.NotificationRepository
	emailService EmailService
	userRepo     repositories.UserRepository
}

func NewNotificationService(notifRepo repositories.NotificationRepository, emailService EmailService, userRepo repositories.UserRepository) NotificationService {
	return &notificationService{
		notifRepo:    notifRepo,
		emailService: emailService,
		userRepo:     userRepo,
	}
}

func (s *notificationService) CreateNotification(userID uuid.UUID, title, message, link, notifType string) {
	// 1. Buat notifikasi in-app
	inAppNotif := &models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Link:    &link,
		Type:    notifType,
	}
	if err := s.notifRepo.Create(inAppNotif); err != nil {
		log.Printf("Failed to create in-app notification for user %s: %v", userID, err)
	}

	// 2. Kirim notifikasi email di background (goroutine)
	go func() {
		user, err := s.userRepo.FindByID(userID.String())
		if err != nil {
			log.Printf("Failed to send email: could not find user %s", userID)
			return
		}
		err = s.emailService.SendEmail(user.Email, user.Name, title, message)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", user.Email, err)
		}
	}()
}