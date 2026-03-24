package model

import "strings"

const (
	TokenDonationStatusPending  = "pending"
	TokenDonationStatusApproved = "approved"
	TokenDonationStatusRejected = "rejected"
)

type TokenDonation struct {
	Id                 int     `json:"id"`
	UserId             int     `json:"user_id" gorm:"index;not null"`
	Type               int     `json:"type" gorm:"index;not null"`
	Name               string  `json:"name" gorm:"type:varchar(128);not null"`
	Key                string  `json:"key" gorm:"type:text;not null"`
	BaseURL            *string `json:"base_url" gorm:"column:base_url;type:varchar(255)"`
	Models             string  `json:"models" gorm:"type:text;not null"`
	Group              string  `json:"group" gorm:"type:varchar(64);default:'default'"`
	Remark             *string `json:"remark" gorm:"type:varchar(255)"`
	OpenAIOrganization *string `json:"openai_organization" gorm:"type:varchar(255)"`
	Status             string  `json:"status" gorm:"type:varchar(32);index;default:'pending'"`
	ReviewNote         *string `json:"review_note" gorm:"type:varchar(255)"`
	ChannelId          *int    `json:"channel_id" gorm:"index"`
	CreatedTime        int64   `json:"created_time" gorm:"bigint;index"`
	ReviewedTime       int64   `json:"reviewed_time" gorm:"bigint"`
	ReviewedBy         int     `json:"reviewed_by" gorm:"index"`
}

type TokenDonationWithUser struct {
	TokenDonation
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (donation *TokenDonation) Insert() error {
	if donation.Group == "" {
		donation.Group = "default"
	}
	return DB.Create(donation).Error
}

func (donation *TokenDonation) Save() error {
	return DB.Save(donation).Error
}

func GetTokenDonationById(id int) (*TokenDonation, error) {
	donation := &TokenDonation{}
	err := DB.First(donation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return donation, nil
}

func GetUserTokenDonations(userId int) ([]*TokenDonation, error) {
	donations := make([]*TokenDonation, 0)
	err := DB.Where("user_id = ?", userId).Order("id desc").Find(&donations).Error
	return donations, err
}

func GetTokenDonationsWithUsers(status string) ([]*TokenDonationWithUser, error) {
	donations := make([]*TokenDonationWithUser, 0)
	query := DB.Table("token_donations").
		Select("token_donations.*, users.username, users.email").
		Joins("left join users on users.id = token_donations.user_id").
		Order("token_donations.id desc")
	status = strings.TrimSpace(status)
	if status != "" && status != "all" {
		query = query.Where("token_donations.status = ?", status)
	}
	err := query.Scan(&donations).Error
	return donations, err
}
