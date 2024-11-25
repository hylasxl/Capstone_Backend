package main

import (
	"Capstone_Go_gRPC/configs"
	"Capstone_Go_gRPC/pkg/models"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	_ = godotenv.Load()
	db, _ := configs.ConnectMySQL()
	err := db.AutoMigrate(
		&models.Account{},
		&models.AccountAvatar{},
		&models.AccountAvatarHistory{},
		&models.AccountChangeNameHistory{},
		&models.AccountInfo{},
		&models.AccountOtpInputs{},
		&models.AccountRole{},
		&models.BackupFile{},
		&models.BackupHistory{},
		&models.DataFieldIndex{},
		&models.DataFieldPrivacy{},
		&models.FriendList{},
		&models.FriendFollow{},
		&models.FriendBlock{},
		&models.FriendListRequest{},
		&models.GBannedWord{},
		&models.OnlineHistory{},
		&models.Permission{},
		&models.PermissionByAccountRole{},
		&models.Post{},
		&models.PostComment{},
		&models.PostCommentEditHistory{},
		&models.PostContentEditHistory{},
		&models.PostMultiMedia{},
		&models.PostMultiMediaReaction{},
		&models.PostReaction{},
		&models.PostMultiMediaComment{},
		&models.PostReported{},
		&models.PostTagFriend{},
		&models.OTPRetakePassword{},
	)
	if err != nil {
		print("Error when migrate model")
		return
	} else {
		log.Default().Print("Migrate MySQL successfully")
	}
}
