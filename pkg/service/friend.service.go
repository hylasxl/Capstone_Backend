package service

import (
	"Capstone_Go_gRPC/pkg/models"
	"Capstone_Go_gRPC/pkg/pb/friendpb"
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
)

type FriendServiceServer struct {
	friendpb.UnimplementedFriendServiceServer
	DB *gorm.DB
}

func (svc *FriendServiceServer) GetAccountListFriend(ctx context.Context, req *friendpb.GetFriendListRequest) (*friendpb.GetFriendListResponse, error) {
	accountID, err := strconv.ParseUint(req.AccountId, 10, 64)
	if err != nil {
		return &friendpb.GetFriendListResponse{
			Error:     "Invalid account ID",
			ErrorCode: "INVALID_ACCOUNT_ID",
		}, nil
	}

	var friendList []models.FriendList
	if err := svc.DB.Where("(first_account_id = ? OR second_account_id = ?) AND is_valid = true", accountID, accountID).Find(&friendList).Error; err != nil {
		return &friendpb.GetFriendListResponse{
			Error:     "Failed to fetch friend list",
			ErrorCode: "DATABASE_ERROR",
		}, nil
	}

	var friends []*friendpb.BasicFriendData
	for _, friend := range friendList {
		friendAccountID := friend.FirstAccountID
		if friendAccountID == uint(accountID) {
			friendAccountID = friend.SecondAccountID
		}

		var accountInfo models.AccountInfo
		if err := svc.DB.Preload("Avatar").Where("account_id = ?", friendAccountID).First(&accountInfo).Error; err != nil {
			continue
		}

		friends = append(friends, &friendpb.BasicFriendData{
			AccountId:       strconv.FormatUint(uint64(accountInfo.AccountID), 10),
			FirstName:       accountInfo.FirstName,
			LastName:        accountInfo.LastName,
			NameDisplayType: string(accountInfo.NameDisplayType),
			AvatarURL:       accountInfo.Avatar.AvatarURL,
		})
	}

	return &friendpb.GetFriendListResponse{
		Friends:   friends,
		Error:     "",
		ErrorCode: "",
	}, nil
}

func (svc *FriendServiceServer) SendFriendList(ctx context.Context, req *friendpb.SendFriendListRequest) (*friendpb.SendFriendListResponse, error) {
	firstAccountID, err := strconv.ParseUint(req.FirstAccountId, 10, 64)
	if err != nil {
		return &friendpb.SendFriendListResponse{
			Error:     "Invalid first account ID",
			ErrorCode: "INVALID_FIRST_ACCOUNT_ID",
		}, nil
	}
	secondAccountID, err := strconv.ParseUint(req.SecondAccountId, 10, 64)
	if err != nil {
		return &friendpb.SendFriendListResponse{
			Error:     "Invalid second account ID",
			ErrorCode: "INVALID_SECOND_ACCOUNT_ID",
		}, nil
	}

	if firstAccountID == secondAccountID {
		return &friendpb.SendFriendListResponse{
			Error:     "Cannot send request to self",
			ErrorCode: "CANNOT_SEND_SELF_REQUEST",
		}, nil
	}

	tx := svc.DB.Begin()

	var existingFriend models.FriendList
	if err := tx.Where(
		"(first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)",
		firstAccountID, secondAccountID, secondAccountID, firstAccountID,
	).First(&existingFriend).Error; err == nil {
		tx.Rollback()
		return &friendpb.SendFriendListResponse{
			Error:     "Accounts are already friends",
			ErrorCode: "ALREADY_FRIENDS",
		}, nil
	}

	var blockedFriend models.FriendBlock
	if err := tx.Where(
		"((first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)) AND is_blocked = true",
		firstAccountID, secondAccountID, secondAccountID, firstAccountID,
	).First(&blockedFriend).Error; err == nil {
		tx.Rollback()
		return &friendpb.SendFriendListResponse{
			Error:     "Friend request blocked",
			ErrorCode: "BLOCKED_RELATIONSHIP",
		}, nil
	}

	var existingRequest models.FriendListRequest
	if err := tx.Where(
		"((sender_account_id = ? AND receiver_account_id = ?) OR (sender_account_id = ? AND receiver_account_id = ?)) AND request_status = 'pending' AND is_recalled = false",
		firstAccountID, secondAccountID, secondAccountID, firstAccountID,
	).First(&existingRequest).Error; err == nil {
		tx.Rollback()
		return &friendpb.SendFriendListResponse{
			Error:     "Friend request already sent",
			ErrorCode: "PENDING_REQUEST_EXISTS",
		}, nil
	}

	newRequest := models.FriendListRequest{
		SenderAccountID:   uint(firstAccountID),
		ReceiverAccountID: uint(secondAccountID),
	}

	if err := tx.Create(&newRequest).Error; err != nil {
		tx.Rollback()
		return &friendpb.SendFriendListResponse{
			Error:     "Failed to send friend request",
			ErrorCode: "DATABASE_ERROR",
		}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &friendpb.SendFriendListResponse{
			Error:     "Transaction commit failed",
			ErrorCode: "TRANSACTION_ERROR",
		}, nil
	}

	return &friendpb.SendFriendListResponse{
		Error:     "",
		ErrorCode: "",
	}, nil
}

func (svc *FriendServiceServer) ResolveFriendRequestAction(ctx context.Context, req *friendpb.FriendRequestActionRequest) (*friendpb.FriendRequestActionResponse, error) {
	if req.Action != "accept" && req.Action != "reject" {
		return &friendpb.FriendRequestActionResponse{
			Error:     "Undefined action",
			ErrorCode: "UNDEFINED_ACTION",
		}, nil
	}

	receiverId, err := strconv.ParseUint(req.ReceiverId, 10, 64)
	if err != nil {
		return &friendpb.FriendRequestActionResponse{
			Error:     "Invalid receiver account ID",
			ErrorCode: "INVALID_RECEIVER_ACCOUNT_ID",
		}, nil
	}

	requestId, err := strconv.ParseUint(req.RequestId, 10, 64)
	if err != nil {
		return &friendpb.FriendRequestActionResponse{
			Error:     "Invalid friend request ID",
			ErrorCode: "INVALID_FRIEND_REQUEST_ID",
		}, nil
	}

	tx := svc.DB.Begin()

	var friendRequest models.FriendListRequest
	if err := tx.Where(
		"receiver_account_id = ? AND id = ? AND is_recalled = false AND request_status = 'pending'",
		receiverId, requestId,
	).First(&friendRequest).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &friendpb.FriendRequestActionResponse{
				Error:     "Request not found",
				ErrorCode: "REQUEST_NOT_FOUND",
			}, nil
		}
	}

	switch req.Action {
	case "accept":
		{
			friendList := &models.FriendList{}
			if err := tx.Where(
				"(first_account_id = ? AND second_account_id = ?) OR (first_account_id = ? AND second_account_id = ?)",
				friendRequest.SenderAccountID, uint(receiverId), uint(receiverId), friendRequest.SenderAccountID).
				First(friendList).Error; err != nil {

				if errors.Is(err, gorm.ErrRecordNotFound) {
					friendList = &models.FriendList{
						FirstAccountID:  uint(receiverId),
						SecondAccountID: friendRequest.SenderAccountID,
					}
					if err := tx.Create(friendList).Error; err != nil {
						tx.Rollback()
						return &friendpb.FriendRequestActionResponse{
							Error:     "Error creating friendship",
							ErrorCode: "CREATE_FRIENDSHIP_ERROR",
						}, nil
					}
				} else {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error querying friendship relationship",
						ErrorCode: "QUERY_FRIENDSHIP_ERROR",
					}, nil
				}
			} else {
				if err := tx.Model(friendList).Update("is_valid", true).Error; err != nil {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error updating friendship relationship",
						ErrorCode: "UPDATE_FRIENDSHIP_ERROR",
					}, nil
				}
			}

			friendFollow := &models.FriendFollow{
				FromAccountID: friendRequest.SenderAccountID,
				ToAccountID:   uint(receiverId),
			}
			if err := tx.Where(
				"from_account_id = ? AND to_account_id = ?",
				friendRequest.SenderAccountID, uint(receiverId)).
				First(friendFollow).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(friendFollow).Error; err != nil {
						tx.Rollback()
						return &friendpb.FriendRequestActionResponse{
							Error:     "Error creating follow relationship",
							ErrorCode: "CREATE_FOLLOW_ERROR",
						}, nil
					}
				} else {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error querying follow relationship",
						ErrorCode: "QUERY_FOLLOW_ERROR",
					}, nil
				}
			} else {
				if err := tx.Model(friendFollow).Update("is_followed", true).Error; err != nil {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error updating follow relationship",
						ErrorCode: "UPDATE_FOLLOW_ERROR",
					}, nil
				}
			}

			friendFollowReversed := &models.FriendFollow{
				FromAccountID: uint(receiverId),
				ToAccountID:   friendRequest.SenderAccountID,
			}
			if err := tx.Where(
				"from_account_id = ? AND to_account_id = ?",
				uint(receiverId), friendRequest.SenderAccountID).
				First(friendFollowReversed).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(friendFollowReversed).Error; err != nil {
						tx.Rollback()
						return &friendpb.FriendRequestActionResponse{
							Error:     "Error creating reversed follow relationship",
							ErrorCode: "CREATE_REVERSED_FOLLOW_ERROR",
						}, nil
					}
				} else {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error querying reversed follow relationship",
						ErrorCode: "QUERY_REVERSED_FOLLOW_ERROR",
					}, nil
				}
			} else {
				if err := tx.Model(friendFollowReversed).Update("is_followed", true).Error; err != nil {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error updating reversed follow relationship",
						ErrorCode: "UPDATE_REVERSED_FOLLOW_ERROR",
					}, nil
				}
			}

			friendBlocked := &models.FriendBlock{}
			if err := tx.Where(
				"(first_account_id = ? AND second_account_id = ?)",
				friendRequest.SenderAccountID, uint(receiverId)).
				First(friendBlocked).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					friendBlocked.FirstAccountID = friendRequest.SenderAccountID
					friendBlocked.SecondAccountID = uint(receiverId)
					if err := tx.Create(friendBlocked).Error; err != nil {
						tx.Rollback()
						return &friendpb.FriendRequestActionResponse{
							Error:     "Error creating block relationship",
							ErrorCode: "CREATE_BLOCK_ERROR",
						}, nil
					}
				} else {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error querying block relationship",
						ErrorCode: "QUERY_BLOCK_ERROR",
					}, nil
				}
			} else {
				if err := tx.Model(friendBlocked).Update("is_blocked", false).Error; err != nil {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error updating block relationship",
						ErrorCode: "UPDATE_BLOCK_ERROR",
					}, nil
				}
			}

			reserverFriendBlock := &models.FriendBlock{}
			if err := tx.Where(
				"(first_account_id = ? AND second_account_id = ?)",
				uint(receiverId), friendRequest.SenderAccountID).
				First(reserverFriendBlock).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					reserverFriendBlock.FirstAccountID = uint(receiverId)
					reserverFriendBlock.SecondAccountID = friendRequest.SenderAccountID
					if err := tx.Create(reserverFriendBlock).Error; err != nil {
						tx.Rollback()
						return &friendpb.FriendRequestActionResponse{
							Error:     "Error creating block relationship",
							ErrorCode: "CREATE_BLOCK_ERROR",
						}, nil
					}
				} else {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error querying block relationship",
						ErrorCode: "QUERY_BLOCK_ERROR",
					}, nil
				}
			} else {
				if err := tx.Model(reserverFriendBlock).Update("is_blocked", false).Error; err != nil {
					tx.Rollback()
					return &friendpb.FriendRequestActionResponse{
						Error:     "Error updating block relationship",
						ErrorCode: "UPDATE_BLOCK_ERROR",
					}, nil
				}
			}

			if err := tx.Model(&friendRequest).Update("request_status", "approved").Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendRequestActionResponse{
					Error:     "Error updating request status",
					ErrorCode: "UPDATE_REQUEST_STATUS_ERROR",
				}, nil
			}

			tx.Commit()

			return &friendpb.FriendRequestActionResponse{
				Error:     "",
				ErrorCode: "",
			}, nil
		}

	case "reject":
		{
			if err := tx.Model(&friendRequest).Update("request_status", "rejected").Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendRequestActionResponse{
					Error:     "Error updating request status",
					ErrorCode: "UPDATE_REQUEST_STATUS_ERROR",
				}, nil
			}
		}
	}

	tx.Commit()
	return &friendpb.FriendRequestActionResponse{
		Error:     "",
		ErrorCode: "",
	}, nil
}

func (svc *FriendServiceServer) RecallFriendRequest(ctx context.Context, req *friendpb.FriendRequestRecallRequest) (*friendpb.FriendRequestRecallResponse, error) {

	senderId, err := strconv.ParseUint(req.SenderId, 10, 64)
	if err != nil {
		return &friendpb.FriendRequestRecallResponse{
			Error:     "Invalid sender account ID",
			ErrorCode: "INVALID_SENDER_ACCOUNT_ID",
		}, nil
	}
	requestId, err := strconv.ParseUint(req.RequestId, 10, 64)
	if err != nil {
		return &friendpb.FriendRequestRecallResponse{
			Error:     "Invalid request ID",
			ErrorCode: "INVALID_REQUEST_ID",
		}, nil
	}

	tx := svc.DB.Begin()

	var friendRequest models.FriendListRequest
	if err := svc.DB.Where(
		"id = ? AND sender_account_id = ?  AND request_status = 'pending' AND is_recalled = false",
		requestId, senderId,
	).First(&friendRequest).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &friendpb.FriendRequestRecallResponse{
				Error:     "Request not found",
				ErrorCode: "REQUEST_NOT_FOUND",
			}, nil
		}
	}

	if err := tx.Model(friendRequest).Update("is_recalled", true).Error; err != nil {
		tx.Rollback()
		return &friendpb.FriendRequestRecallResponse{
			Error:     "Error recall request",
			ErrorCode: "ERROR_RECALL_REQUEST",
		}, nil
	}

	tx.Commit()

	return &friendpb.FriendRequestRecallResponse{
		Error:     "",
		ErrorCode: "",
	}, nil
}

func (svc *FriendServiceServer) FollowFriend(ctx context.Context, req *friendpb.FriendFollowRequest) (*friendpb.FriendFollowResponse, error) {
	fromId, err := strconv.ParseUint(req.FromAccountId, 10, 64)
	if err != nil {
		return &friendpb.FriendFollowResponse{
			Error:     "Invalid sender account ID",
			ErrorCode: "INVALID_SENDER_ACCOUNT_ID",
		}, nil
	}

	toId, err := strconv.ParseUint(req.ToAccountId, 10, 64)
	if err != nil {
		return &friendpb.FriendFollowResponse{
			Error:     "Invalid target account ID",
			ErrorCode: "INVALID_TARGET_ACCOUNT_ID",
		}, nil
	}

	tx := svc.DB.Begin()

	blockFriend := &models.FriendBlock{}

	if err := tx.Model(blockFriend).Where("(first_account_id = ? AND second_account_id = ?)", uint(fromId), uint(toId)).First(blockFriend).Error; err != nil {
	} else {
		if blockFriend.IsBlocked {
			tx.Rollback()
			return &friendpb.FriendFollowResponse{
				Error:     "Target account is blocked",
				ErrorCode: "TARGET_ACCOUNT_BLOCKED",
			}, nil
		}
	}

	blockedFriend := &models.FriendBlock{}
	if err := tx.Model(blockedFriend).Where("(first_account_id = ? AND second_account_id = ?)", uint(toId), uint(fromId)).First(blockedFriend).Error; err != nil {
	} else {
		if blockedFriend.IsBlocked {
			tx.Rollback()
			return &friendpb.FriendFollowResponse{
				Error:     "Sender account is blocked",
				ErrorCode: "SENDER_ACCOUNT_BLOCKED",
			}, nil
		}
	}

	friendFollow := &models.FriendFollow{}
	err = tx.Where(
		"from_account_id = ? AND to_account_id = ?",
		fromId, toId,
	).First(friendFollow).Error

	switch req.Action {
	case "follow":
		if errors.Is(err, gorm.ErrRecordNotFound) {
			friendFollow = &models.FriendFollow{
				FromAccountID: uint(fromId),
				ToAccountID:   uint(toId),
				IsFollowed:    true,
			}
			if err := tx.Create(friendFollow).Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendFollowResponse{
					Error:     "Error creating follow relationship",
					ErrorCode: "CREATE_FOLLOW_ERROR",
				}, nil
			}
		} else if err == nil {
			if err := tx.Model(friendFollow).Update("is_followed", true).Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendFollowResponse{
					Error:     "Error updating follow relationship",
					ErrorCode: "UPDATE_FOLLOW_ERROR",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friendpb.FriendFollowResponse{
				Error:     "Error querying follow relationship",
				ErrorCode: "QUERY_FOLLOW_ERROR",
			}, nil
		}

	case "unfollow":
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &friendpb.FriendFollowResponse{
				Error:     "No follow relationship exists to unfollow",
				ErrorCode: "UNFOLLOW_NOT_FOUND_ERROR",
			}, nil
		} else if err == nil {
			if err := tx.Model(friendFollow).Update("is_followed", false).Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendFollowResponse{
					Error:     "Error updating follow relationship",
					ErrorCode: "UPDATE_FOLLOW_ERROR",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friendpb.FriendFollowResponse{
				Error:     "Error querying follow relationship",
				ErrorCode: "QUERY_FOLLOW_ERROR",
			}, nil
		}

	default:
		tx.Rollback()
		return &friendpb.FriendFollowResponse{
			Error:     "Invalid action specified",
			ErrorCode: "INVALID_ACTION_ERROR",
		}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &friendpb.FriendFollowResponse{
			Error:     "Transaction commit failed",
			ErrorCode: "TRANSACTION_COMMIT_ERROR",
		}, nil
	}

	return &friendpb.FriendFollowResponse{
		Error:     "",
		ErrorCode: "",
	}, nil
}

func (svc *FriendServiceServer) BlockFriend(ctx context.Context, req *friendpb.FriendBlockRequest) (*friendpb.FriendBlockResponse, error) {
	firstId, err := strconv.ParseUint(req.FirstAccountId, 10, 64)
	if err != nil {
		return &friendpb.FriendBlockResponse{
			Error:     "Invalid first account ID",
			ErrorCode: "INVALID_FIRST_ACCOUNT_ID",
		}, nil
	}

	secondId, err := strconv.ParseUint(req.SecondAccountId, 10, 64)
	if err != nil {
		return &friendpb.FriendBlockResponse{
			Error:     "Invalid second account ID",
			ErrorCode: "INVALID_SECOND_ACCOUNT_ID",
		}, nil
	}

	var primaryId, secondaryId uint64
	if firstId < secondId {
		primaryId, secondaryId = firstId, secondId
	} else {
		primaryId, secondaryId = secondId, firstId
	}

	tx := svc.DB.Begin()

	friendBlock := &models.FriendBlock{}
	err = tx.Where(
		"(first_account_id = ? AND second_account_id = ?)",
		primaryId, secondaryId,
	).First(friendBlock).Error

	switch req.Action {
	case "block":
		if errors.Is(err, gorm.ErrRecordNotFound) {
			friendBlock = &models.FriendBlock{
				FirstAccountID:  uint(primaryId),
				SecondAccountID: uint(secondaryId),
				IsBlocked:       true,
			}
			if err := tx.Create(friendBlock).Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendBlockResponse{
					Error:     "Error creating block relationship",
					ErrorCode: "CREATE_BLOCK_ERROR",
				}, nil
			}
		} else if err == nil {
			if err := tx.Model(friendBlock).Update("is_blocked", true).Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendBlockResponse{
					Error:     "Error updating block relationship",
					ErrorCode: "UPDATE_BLOCK_ERROR",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friendpb.FriendBlockResponse{
				Error:     "Error querying block relationship",
				ErrorCode: "QUERY_BLOCK_ERROR",
			}, nil
		}

	case "unblock":
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return &friendpb.FriendBlockResponse{
				Error:     "No block relationship exists to unblock",
				ErrorCode: "UNBLOCK_NOT_FOUND_ERROR",
			}, nil
		} else if err == nil {
			if err := tx.Model(friendBlock).Update("is_blocked", false).Error; err != nil {
				tx.Rollback()
				return &friendpb.FriendBlockResponse{
					Error:     "Error updating block relationship",
					ErrorCode: "UPDATE_BLOCK_ERROR",
				}, nil
			}
		} else {
			tx.Rollback()
			return &friendpb.FriendBlockResponse{
				Error:     "Error querying block relationship",
				ErrorCode: "QUERY_BLOCK_ERROR",
			}, nil
		}

	default:
		tx.Rollback()
		return &friendpb.FriendBlockResponse{
			Error:     "Invalid action specified",
			ErrorCode: "INVALID_ACTION_ERROR",
		}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &friendpb.FriendBlockResponse{
			Error:     "Transaction commit failed",
			ErrorCode: "TRANSACTION_COMMIT_ERROR",
		}, nil
	}

	return &friendpb.FriendBlockResponse{
		Error:     "",
		ErrorCode: "",
	}, nil
}
