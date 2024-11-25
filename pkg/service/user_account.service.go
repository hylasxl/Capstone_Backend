package service

import (
	"Capstone_Go_gRPC/pkg/models"
	"Capstone_Go_gRPC/pkg/pb/userAccountpb"
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"strconv"
)

type UserAccountServiceServer struct {
	userAccountpb.UnimplementedUserAccountServer
	DB *gorm.DB
}

func (svc *UserAccountServiceServer) GetAccountInfo(ctx context.Context, req *userAccountpb.GetAccountInfoRequest) (*userAccountpb.GetAccountInfoResponse, error) {
	res := &userAccountpb.GetAccountInfoResponse{}

	var acc models.Account
	if err := svc.DB.First(&acc, "id = ?", req.AccountId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Error = "Account not found"
			res.ErrorCode = "ACNF"
			return res, nil
		}
		res.Error = "Internal server error"
		res.ErrorCode = "ISE"
		return res, err
	}
	res.Account = &userAccountpb.Account{
		Id:            strconv.Itoa(int(acc.ID)),
		Username:      acc.Username,
		AccountRoleId: uint32(acc.AccountRoleID),
		CreatedBy:     acc.AccountCreatedByMethod.ToProto(),
		IsBanned:      acc.IsBanned,
		IsRestricted:  acc.IsRestricted,
		IsSelfDeleted: acc.IsSelfDeleted,
		CreatedAt:     timestamppb.New(acc.CreatedAt),
	}

	var accInfo models.AccountInfo
	if err := svc.DB.First(&accInfo, "account_id = ?", req.AccountId).Error; err == nil {
		res.AccountInfo = &userAccountpb.AccountInfo{
			Id:              strconv.Itoa(int(accInfo.ID)),
			AccountId:       strconv.Itoa(int(accInfo.AccountID)),
			AvatarId:        strconv.Itoa(int(accInfo.AvatarID)),
			FirstName:       accInfo.FirstName,
			LastName:        accInfo.LastName,
			DateOfBirth:     timestamppb.New(accInfo.DateOfBirth),
			Gender:          accInfo.Gender.ToProto(),
			MaritalStatus:   accInfo.MaritalStatus.ToProto(),
			PhoneNumber:     accInfo.PhoneNumber,
			Email:           accInfo.Email,
			NameDisplayType: accInfo.NameDisplayType.ToProto(),
			CreatedAt:       timestamppb.New(accInfo.CreatedAt),
		}
	}

	var accAvt models.AccountAvatar
	if err := svc.DB.First(&accAvt, "account_id = ?", req.AccountId).Error; err == nil {
		res.Avatar = &userAccountpb.AccountAvatar{
			Id:             strconv.Itoa(int(accAvt.ID)),
			AvatarUrl:      accAvt.AvatarURL,
			IsInUse:        accAvt.IsInUsed,
			IsDeleted:      accAvt.IsDeleted,
			IsUsingDefault: accAvt.IsUsingDefault,
			CreatedAt:      timestamppb.New(accAvt.CreatedAt),
		}
	}
	return res, nil
}
