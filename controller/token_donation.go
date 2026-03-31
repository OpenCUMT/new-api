package controller

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var supportedTokenDonationChannelTypes = map[int]struct{}{
	constant.ChannelTypeOpenAI: {},
}

type tokenDonationCreateRequest struct {
	Type        int    `json:"type"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	BaseURL     string `json:"base_url"`
	TokenSource string `json:"token_source"`
	Models      string `json:"models"`
	Group       string `json:"group"`
	Remark      string `json:"remark"`
}

type tokenDonationResponse struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	Type        int    `json:"type"`
	TypeName    string `json:"type_name"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	BaseURL     string `json:"base_url,omitempty"`
	TokenSource string `json:"token_source"`
	Models      string `json:"models"`
	Group       string `json:"group"`
	Remark      string `json:"remark,omitempty"`
	Status      string `json:"status"`
	ReviewNote  string `json:"review_note,omitempty"`
	ChannelId   *int   `json:"channel_id,omitempty"`
	CreatedTime int64  `json:"created_time"`
	ReviewedTime int64 `json:"reviewed_time"`
	ReviewedBy  int    `json:"reviewed_by"`
}

type tokenDonationRankingItem struct {
	Rank              int    `json:"rank"`
	UserId            int    `json:"user_id"`
	Username          string `json:"username"`
	DonationCount     int    `json:"donation_count"`
	LatestDonationTime int64 `json:"latest_donation_time"`
}

func isSupportedTokenDonationChannelType(channelType int) bool {
	_, ok := supportedTokenDonationChannelTypes[channelType]
	return ok
}

func normalizeCSV(input string) string {
	items := strings.Split(input, ",")
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return strings.Join(result, ",")
}

func trimOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func truncateString(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validateTokenDonationBaseURL(baseURL string) error {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil
	}
	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("base URL only supports http or https")
	}
	return nil
}

func validateTokenDonationRequest(req *tokenDonationCreateRequest) error {
	if req.Type == 0 {
		req.Type = constant.ChannelTypeOpenAI
	}
	if req.Type != constant.ChannelTypeOpenAI {
		return errors.New("token donation currently only supports OpenAI channel type")
	}
	if !isSupportedTokenDonationChannelType(req.Type) {
		return errors.New("this channel type does not support token donation")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Key = strings.TrimSpace(req.Key)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.TokenSource = strings.TrimSpace(req.TokenSource)
	req.Models = normalizeCSV(req.Models)
	req.Group = normalizeCSV(req.Group)
	req.Remark = strings.TrimSpace(req.Remark)

	if req.Name == "" {
		return errors.New("channel name is required")
	}
	if req.Key == "" {
		return errors.New("token is required")
	}
	if req.BaseURL == "" {
		return errors.New("base URL is required")
	}
	if req.TokenSource == "" {
		return errors.New("token source is required")
	}
	if req.Models == "" {
		return errors.New("models are required, separate multiple models with commas")
	}
	if req.Group == "" {
		return errors.New("group is required")
	}
	if len(req.Name) > 128 {
		return errors.New("channel name is too long")
	}
	if len(req.TokenSource) > 255 {
		return errors.New("token source is too long")
	}
	if len(req.Remark) > 255 {
		return errors.New("remark is too long")
	}
	if err := validateTokenDonationBaseURL(req.BaseURL); err != nil {
		return err
	}

	channel := &model.Channel{
		Type:   req.Type,
		Name:   req.Name,
		Key:    req.Key,
		Models: req.Models,
		Group:  req.Group,
	}
	if req.BaseURL != "" {
		channel.BaseURL = trimOptionalString(req.BaseURL)
	}
	if err := validateChannel(channel, true); err != nil {
		return err
	}
	return nil
}

func buildTokenDonationResponse(donation *model.TokenDonation, username string, email string) tokenDonationResponse {
	baseURL := ""
	if donation.BaseURL != nil {
		baseURL = *donation.BaseURL
	}
	remark := ""
	if donation.Remark != nil {
		remark = *donation.Remark
	}
	reviewNote := ""
	if donation.ReviewNote != nil {
		reviewNote = *donation.ReviewNote
	}
	return tokenDonationResponse{
		Id:           donation.Id,
		UserId:       donation.UserId,
		Username:     username,
		Email:        email,
		Type:         donation.Type,
		TypeName:     constant.GetChannelTypeName(donation.Type),
		Name:         donation.Name,
		Key:          model.MaskTokenKey(donation.Key),
		BaseURL:      baseURL,
		TokenSource:  donation.TokenSource,
		Models:       donation.Models,
		Group:        donation.Group,
		Remark:       remark,
		Status:       donation.Status,
		ReviewNote:   reviewNote,
		ChannelId:    donation.ChannelId,
		CreatedTime:  donation.CreatedTime,
		ReviewedTime: donation.ReviewedTime,
		ReviewedBy:   donation.ReviewedBy,
	}
}

func buildChannelFromTokenDonation(donation *model.TokenDonation) *model.Channel {
	channel := &model.Channel{
		Type:        donation.Type,
		Name:        donation.Name,
		Key:         donation.Key,
		Status:      common.ChannelStatusEnabled,
		CreatedTime: common.GetTimestamp(),
		Models:      donation.Models,
		Group:       donation.Group,
	}
	if donation.BaseURL != nil && strings.TrimSpace(*donation.BaseURL) != "" {
		baseURL := strings.TrimSpace(*donation.BaseURL)
		channel.BaseURL = &baseURL
	}
	tag := "token-donation"
	channel.Tag = &tag
	remark := fmt.Sprintf("[Donation #%d][User #%d][Source: %s] %s", donation.Id, donation.UserId, strings.TrimSpace(donation.TokenSource), strings.TrimSpace(stringOrEmpty(donation.Remark)))
	remark = truncateString(strings.TrimSpace(remark), 255)
	if remark != "" {
		channel.Remark = &remark
	}
	return channel
}

func CreateTokenDonation(c *gin.Context) {
	userId := c.GetInt("id")
	var req tokenDonationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := validateTokenDonationRequest(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	donation := &model.TokenDonation{
		UserId:      userId,
		Type:        req.Type,
		Name:        req.Name,
		Key:         req.Key,
		BaseURL:     trimOptionalString(req.BaseURL),
		TokenSource: req.TokenSource,
		Models:      req.Models,
		Group:       req.Group,
		Remark:      trimOptionalString(req.Remark),
		Status:      model.TokenDonationStatusPending,
		CreatedTime: common.GetTimestamp(),
	}
	if err := donation.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, buildTokenDonationResponse(donation, "", ""))
}

func GetSelfTokenDonations(c *gin.Context) {
	userId := c.GetInt("id")
	donations, err := model.GetUserTokenDonations(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]tokenDonationResponse, 0, len(donations))
	for _, donation := range donations {
		items = append(items, buildTokenDonationResponse(donation, "", ""))
	}
	common.ApiSuccess(c, items)
}

func GetAllTokenDonations(c *gin.Context) {
	donations, err := model.GetTokenDonationsWithUsers(c.Query("status"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]tokenDonationResponse, 0, len(donations))
	for _, donation := range donations {
		items = append(items, buildTokenDonationResponse(&donation.TokenDonation, donation.Username, donation.Email))
	}
	common.ApiSuccess(c, items)
}

func GetTokenDonationRanking(c *gin.Context) {
	donations, err := model.GetTokenDonationsWithUsers(model.TokenDonationStatusApproved)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	userRankMap := make(map[int]*tokenDonationRankingItem)
	for _, donation := range donations {
		if donation == nil || donation.UserId <= 0 {
			continue
		}

		item, ok := userRankMap[donation.UserId]
		if !ok {
			username := strings.TrimSpace(donation.Username)
			if username == "" {
				username = fmt.Sprintf("User #%d", donation.UserId)
			}
			item = &tokenDonationRankingItem{
				UserId:   donation.UserId,
				Username: username,
			}
			userRankMap[donation.UserId] = item
		}

		item.DonationCount++
		latestTime := donation.ReviewedTime
		if latestTime <= 0 {
			latestTime = donation.CreatedTime
		}
		if latestTime > item.LatestDonationTime {
			item.LatestDonationTime = latestTime
		}
	}

	ranking := make([]*tokenDonationRankingItem, 0, len(userRankMap))
	for _, item := range userRankMap {
		if item.DonationCount > 0 {
			ranking = append(ranking, item)
		}
	}

	sort.SliceStable(ranking, func(i, j int) bool {
		if ranking[i].DonationCount != ranking[j].DonationCount {
			return ranking[i].DonationCount > ranking[j].DonationCount
		}
		if ranking[i].LatestDonationTime != ranking[j].LatestDonationTime {
			return ranking[i].LatestDonationTime > ranking[j].LatestDonationTime
		}
		return ranking[i].UserId < ranking[j].UserId
	})

	for idx, item := range ranking {
		item.Rank = idx + 1
	}

	common.ApiSuccess(c, ranking)
}

func ApproveTokenDonation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	adminId := c.GetInt("id")

	donation, err := model.GetTokenDonationById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if donation.Status != model.TokenDonationStatusPending {
		common.ApiError(c, errors.New("this donation has already been reviewed"))
		return
	}

	channel := buildChannelFromTokenDonation(donation)
	if err := validateChannel(channel, true); err != nil {
		common.ApiError(c, err)
		return
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		latestDonation := &model.TokenDonation{}
		if err := tx.First(latestDonation, "id = ?", id).Error; err != nil {
			return err
		}
		if latestDonation.Status != model.TokenDonationStatusPending {
			return errors.New("this donation has already been reviewed")
		}
		if err := tx.Create(channel).Error; err != nil {
			return err
		}
		if err := channel.AddAbilities(tx); err != nil {
			return err
		}
		latestDonation.Status = model.TokenDonationStatusApproved
		latestDonation.ChannelId = &channel.Id
		latestDonation.ReviewedBy = adminId
		latestDonation.ReviewedTime = common.GetTimestamp()
		return tx.Save(latestDonation).Error
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"id":         donation.Id,
		"channel_id": channel.Id,
	})
}

func RejectTokenDonation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	adminId := c.GetInt("id")
	donation, err := model.GetTokenDonationById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if donation.Status != model.TokenDonationStatusPending {
		common.ApiError(c, errors.New("this donation has already been reviewed"))
		return
	}
	donation.Status = model.TokenDonationStatusRejected
	donation.ReviewedBy = adminId
	donation.ReviewedTime = common.GetTimestamp()
	if err := donation.Save(); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": donation.Id})
}
