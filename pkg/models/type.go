package models

import "Capstone_Go_gRPC/pkg/pb/userAccountpb"

type AccountCreatedByMethod string
type Gender string
type MaritalStatus string
type NameDisplayType string
type UploadStatus string
type ChangingType string
type MediaType string
type ReactionType string
type PrivacyStatus string
type ReportResolve string
type NormalStatus string
type OTPStatus string

const (
	Like    ReactionType = "like"
	Love    ReactionType = "love"
	Hate    ReactionType = "hate"
	Dislike ReactionType = "dislike"
	Cry     ReactionType = "cry"
)

const (
	Google AccountCreatedByMethod = "google"
	Normal AccountCreatedByMethod = "normal"
)

const (
	Uploaded UploadStatus = "uploaded"
	Failed   UploadStatus = "failed"
)

const (
	DisplayName ChangingType = "display_name"
	UserName    ChangingType = "user_name"
)

const (
	Male   Gender = "male"
	Female Gender = "female"
	Other  Gender = "other"
)

const (
	Single              MaritalStatus = "single"
	InRelationship      MaritalStatus = "in_a_relationship"
	Engaged             MaritalStatus = "engaged"
	Married             MaritalStatus = "married"
	CivilUnion          MaritalStatus = "in_a_civil_union"
	DomesticPartnership MaritalStatus = "in_a_domestic_partnership"
	OpenRelationship    MaritalStatus = "in_an_open_relationship"
	Complicated         MaritalStatus = "it_complicated"
	Separated           MaritalStatus = "separated"
	Divorced            MaritalStatus = "divorced"
	Widowed             MaritalStatus = "widowed"
)

const (
	FirstNameFirst NameDisplayType = "first_name_first"
	LastNameFirst  NameDisplayType = "last_name_first"
)

const (
	Picture MediaType = "picture"
	Video   MediaType = "video"
)

const (
	Public     PrivacyStatus = "public"
	Private    PrivacyStatus = "private"
	FriendOnly PrivacyStatus = "friend_only"
)

const (
	Pending  NormalStatus = "pending"
	Approved NormalStatus = "approved"
	Rejected NormalStatus = "rejected"
)

const (
	ReportPending ReportResolve = "report_pending"
	DeletePost    ReportResolve = "delete_post"
	ReportSkipped ReportResolve = "report_skipped"
)

const (
	Valid   OTPStatus = "valid"
	Expired OTPStatus = "expired"
	Invalid OTPStatus = "invalid"
)

func (m AccountCreatedByMethod) ToProto() userAccountpb.AccountCreatedByMethod {
	switch m {
	case Google:
		return userAccountpb.AccountCreatedByMethod_ACCOUNT_CREATED_BY_GOOGLE
	case Normal:
		return userAccountpb.AccountCreatedByMethod_ACCOUNT_CREATED_BY_NORMAL
	default:
		return userAccountpb.AccountCreatedByMethod_ACCOUNT_CREATED_BY_NORMAL // Default case
	}
}

// Gender ToProto conversion
func (g Gender) ToProto() userAccountpb.Gender {
	switch g {
	case Male:
		return userAccountpb.Gender_MALE
	case Female:
		return userAccountpb.Gender_FEMALE
	case Other:
		return userAccountpb.Gender_OTHER
	default:
		return userAccountpb.Gender_OTHER // Default case
	}
}

// MaritalStatus ToProto conversion
func (m MaritalStatus) ToProto() userAccountpb.MaritalStatus {
	switch m {
	case Single:
		return userAccountpb.MaritalStatus_SINGLE
	case InRelationship:
		return userAccountpb.MaritalStatus_IN_A_RELATIONSHIP
	case Engaged:
		return userAccountpb.MaritalStatus_ENGAGED
	case Married:
		return userAccountpb.MaritalStatus_MARRIED
	case CivilUnion:
		return userAccountpb.MaritalStatus_IN_A_CIVIL_UNION
	case DomesticPartnership:
		return userAccountpb.MaritalStatus_IN_A_DOMESTIC_PARTNERSHIP
	case OpenRelationship:
		return userAccountpb.MaritalStatus_IN_AN_OPEN_RELATIONSHIP
	case Complicated:
		return userAccountpb.MaritalStatus_ITS_COMPLICATED
	case Separated:
		return userAccountpb.MaritalStatus_SEPARATED
	case Divorced:
		return userAccountpb.MaritalStatus_DIVORCED
	case Widowed:
		return userAccountpb.MaritalStatus_WIDOWED
	default:
		return userAccountpb.MaritalStatus_SINGLE // Default case
	}
}

// NameDisplayType ToProto conversion
func (n NameDisplayType) ToProto() userAccountpb.NameDisplayType {
	switch n {
	case FirstNameFirst:
		return userAccountpb.NameDisplayType_FIRST_NAME_FIRST
	case LastNameFirst:
		return userAccountpb.NameDisplayType_LAST_NAME_FIRST
	default:
		return userAccountpb.NameDisplayType_FIRST_NAME_FIRST // Default case
	}
}
