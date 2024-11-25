package service

import (
	"Capstone_Go_gRPC/configs"
	"Capstone_Go_gRPC/pkg/models"
	"Capstone_Go_gRPC/pkg/pb/authpb"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

var (
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	jwtSecretKey         = []byte(os.Getenv("JWT_SECRET_KEY"))
)

type CustomClaims struct {
	AccountId   uint64   `json:"accountId"`
	Permissions []string `json:"permissions"`
	RoleId      uint64   `json:"roleId"`

	jwt.StandardClaims
}

func (c *CustomClaims) Valid() error {
	return c.StandardClaims.Valid()
}

type AuthServiceServer struct {
	authpb.UnimplementedAuthServiceServer
	DB               *gorm.DB
	CloudinaryClient *configs.CloudinaryService
}

func init() {
	var err error
	accessTokenEnv := os.Getenv("ACCESS_TOKEN_DURATION")
	if accessTokenEnv == "" {
		accessTokenEnv = "15m"
	}
	accessTokenDuration, err = time.ParseDuration(accessTokenEnv)
	if err != nil {
		fmt.Println("Error parsing ACCESS_TOKEN_DURATION:", err)
		accessTokenDuration = 15 * time.Minute
	}

	refreshTokenEnv := os.Getenv("REFRESH_TOKEN_DURATION")
	if refreshTokenEnv == "" {
		refreshTokenEnv = "720h"
	}
	refreshTokenDuration, err = time.ParseDuration(refreshTokenEnv)
	if err != nil {
		fmt.Println("Error parsing REFRESH_TOKEN_DURATION:", err)
		refreshTokenDuration = 30 * 24 * time.Hour
	}
}

func (svc *AuthServiceServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	var acc models.Account
	var permissions []string

	if req.Username == "" {
		return &authpb.LoginResponse{Error: "Username cannot be empty", ErrorCode: "LGE01"}, nil
	}
	if req.Password == "" {
		return &authpb.LoginResponse{Error: "Password cannot be empty", ErrorCode: "LGE02"}, nil
	}

	if err := svc.DB.Where("username = ?", req.Username).First(&acc).Error; err != nil {
		return &authpb.LoginResponse{Error: "The username is not correct", ErrorCode: "LGE03"}, nil
	}

	if acc.IsSelfDeleted {
		return &authpb.LoginResponse{Error: "The account is deleted", ErrorCode: "LGE04"}, nil
	}
	if acc.IsRestricted {
		return &authpb.LoginResponse{Error: "The account is restricted by admin", ErrorCode: "LGE05"}, nil
	}
	if acc.IsBanned {
		return &authpb.LoginResponse{Error: "The account is banned by admin", ErrorCode: "LGE06"}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHashed), []byte(req.Password)); err != nil {
		return &authpb.LoginResponse{Error: "The password is not correct", ErrorCode: "LGE07"}, nil
	}

	if err := svc.DB.Model(&acc).
		Select("permissions.permission_url").
		Joins("JOIN account_roles ON account_roles.id = ?", acc.AccountRoleID).
		Joins("JOIN permission_by_account_roles ON permission_by_account_roles.account_role_id = account_roles.id").
		Joins("JOIN permissions ON permissions.id = permission_by_account_roles.permission_id").
		Scan(&permissions).Error; err != nil {
		return nil, err
	}

	accessToken, err := GenerateAccessToken(permissions, int32(acc.ID), int32(acc.AccountRoleID))
	if err != nil {
		return &authpb.LoginResponse{Error: "Error when generating access token"}, nil
	}
	refreshToken, err := GenerateRefreshToken(int32(acc.ID), int32(acc.AccountRoleID))
	if err != nil {
		return &authpb.LoginResponse{Error: "Error when generating refresh token"}, nil
	}

	return &authpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Claims: &authpb.JWTClaims{
			AccountId:   uint64(acc.ID),
			Permissions: permissions,
			RoleId:      uint64(acc.AccountRoleID),
			Issuer:      "SyncIO_Server",
			Subject:     "Authentication",
			Audience:    "",
			ExpiresAt:   timestamppb.New(time.Now().Add(accessTokenDuration)),
		},
		Error: "",
	}, nil
}

func (svc *AuthServiceServer) Signup(ctx context.Context, req *authpb.SignupRequest) (*authpb.SignUpResponse, error) {

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))

	if req.FirstName == "" || req.LastName == "" || req.Birthday == nil || req.Gender == "" ||
		req.Email == "" || req.Password == "" {
		return &authpb.SignUpResponse{Error: "Missing required fields", ErrorCode: "RE01"}, nil
	}

	tx := svc.DB.Begin()

	if tx.Error != nil {
		return &authpb.SignUpResponse{Error: "Error starting transaction", ErrorCode: "MYSQLDBE01"}, nil
	}

	var existingUsername models.Account
	if err := tx.Where("username = ?", req.Username).First(&existingUsername).Error; err == nil {

		tx.Rollback()
		return &authpb.SignUpResponse{Error: "Username already exists", ErrorCode: "RE02"}, nil
	}

	var existingEmail models.AccountInfo
	if err := tx.Where("email = ?", req.Email).First(&existingEmail).Error; err == nil {

		tx.Rollback()
		return &authpb.SignUpResponse{Error: "Email already exists", ErrorCode: "RE03"}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {

		tx.Rollback()
		return &authpb.SignUpResponse{Error: "Error hashing password", ErrorCode: "SE01"}, nil
	}

	account := models.Account{
		Username:               req.Username,
		PasswordHashed:         string(hashedPassword),
		AccountCreatedByMethod: models.Normal,
		AccountRoleID:          1,
	}

	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()

		return &authpb.SignUpResponse{Error: "Error creating account", ErrorCode: "RE04"}, nil
	}

	avatarUrl := "https://res.cloudinary.com/deb9bbqpg/image/upload/v1731228160/user_profile_image/kmvaozdoe8s9n9jx83gy.png"
	var uploadedAvatarUrl string
	if len(req.Avatar) > 0 {
		uploadedAvatarUrl, err := svc.CloudinaryClient.UploadAvatarImage(req.Avatar)
		if err != nil {

			tx.Rollback()
			return &authpb.SignUpResponse{Error: "Error uploading avatar", ErrorCode: "AUE"}, nil
		}
		avatarUrl = uploadedAvatarUrl
	}

	accountAvatar := models.AccountAvatar{
		AvatarURL: avatarUrl,
		AccountID: account.ID,
	}

	if err := tx.Create(&accountAvatar).Error; err != nil {

		tx.Rollback()
		return &authpb.SignUpResponse{Error: "Error creating account avatar", ErrorCode: "RE05"}, nil
	}

	accountInfo := models.AccountInfo{
		AccountID:     account.ID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DateOfBirth:   req.Birthday.AsTime(),
		Gender:        models.Gender(req.Gender),
		Email:         req.Email,
		PhoneNumber:   req.PhoneNumber,
		MaritalStatus: models.Single,
		AvatarID:      accountAvatar.ID,
	}

	if err := tx.Create(&accountInfo).Error; err != nil {

		tx.Rollback()
		if len(req.Avatar) > 0 {
			err := svc.CloudinaryClient.DeleteFile(uploadedAvatarUrl)
			if err != nil {
				return &authpb.SignUpResponse{Error: "Error deleting avatar from Cloudinary", ErrorCode: "AUE_DEL"}, nil
			}
		}
		return &authpb.SignUpResponse{Error: "Error creating account info", ErrorCode: "RE06"}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &authpb.SignUpResponse{Error: "Error committing transaction", ErrorCode: "MYSQLDBE02"}, nil
	}

	return &authpb.SignUpResponse{Error: ""}, nil
}

func GenerateAccessToken(permissions []string, accountId int32, roleId int32) (string, error) {
	claims := &CustomClaims{
		AccountId:   uint64(accountId),
		Permissions: permissions,
		RoleId:      uint64(roleId),
		StandardClaims: jwt.StandardClaims{
			Issuer:    "SyncIO_Server",
			Subject:   "Authentication",
			Audience:  "", // Set an audience if required
			ExpiresAt: time.Now().Add(accessTokenDuration).Unix(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccessToken, err := accessToken.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	return signedAccessToken, nil
}

func GenerateRefreshToken(accountId int32, roleId int32) (string, error) {
	claims := &CustomClaims{
		AccountId: uint64(accountId),
		RoleId:    uint64(roleId),
		StandardClaims: jwt.StandardClaims{
			Issuer:    "SyncIO_Server",
			Subject:   "Authentication",
			Audience:  "",
			ExpiresAt: time.Now().Add(refreshTokenDuration).Unix(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedRefreshToken, err := refreshToken.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	return signedRefreshToken, nil
}

func (svc *AuthServiceServer) CheckExistingUsername(ctx context.Context, req *authpb.CheckExistingUsernameRequest) (*authpb.CheckExistingUsernameResponse, error) {
	var existingUsername models.Account

	err := svc.DB.Where("username = ?", req.Username).First(&existingUsername).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return &authpb.CheckExistingUsernameResponse{IsExisting: false}, nil
		}
		return nil, err
	}

	return &authpb.CheckExistingUsernameResponse{IsExisting: true}, nil
}

func (svc *AuthServiceServer) CheckExistingEmail(ctx context.Context, req *authpb.CheckExistingEmailRequest) (*authpb.CheckExistingEmailResponse, error) {
	var existingEmail models.AccountInfo

	err := svc.DB.Where("email = ?", req.Email).First(&existingEmail).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return &authpb.CheckExistingEmailResponse{IsExisting: false}, nil
		}
		return nil, err
	}

	return &authpb.CheckExistingEmailResponse{IsExisting: true}, nil
}
