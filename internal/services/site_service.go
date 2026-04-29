package services

import (
	"ai-later-nav/internal/database/repository"
	"ai-later-nav/internal/models"
	"ai-later-nav/internal/utils"
	"strings"
)

type SiteService struct {
	siteRepo *repository.SiteRepository
}

func buildDisplayTags(tags []string) []models.DisplayTag {
	displayTags := make([]models.DisplayTag, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.TrimSpace(tag)
		if normalized == "" {
			continue
		}
		displayTags = append(displayTags, models.DisplayTag{
			Name:  normalized,
			Class: utils.GetTagColorClass(normalized),
		})
	}
	return displayTags
}

func buildSiteDisplay(site models.SiteWithTags, todayUV int64) models.SiteDisplay {
	return models.SiteDisplay{
		Site:        site.Site,
		Tags:        site.Tags,
		DisplayTags: buildDisplayTags(site.Tags),
		Color:       utils.GenerateColorFromName(site.Name),
		Initials:    utils.GetInitialsFromName(site.Name),
		TodayUV:     todayUV,
	}
}

func NewSiteService() *SiteService {
	return &SiteService{
		siteRepo: repository.NewSiteRepository(),
	}
}

func (s *SiteService) Count() (int64, error) {
	return s.siteRepo.CountSites()
}

func (s *SiteService) GetAll() ([]models.SiteDisplay, error) {
	sites, err := s.siteRepo.GetAllWithTags()
	if err != nil {
		return nil, err
	}

	todayUVMap, err := s.siteRepo.GetAllSitesTodayUV()
	if err != nil {
		todayUVMap = make(map[int64]int64)
	}

	var result []models.SiteDisplay
	for _, swt := range sites {
		result = append(result, buildSiteDisplay(swt, todayUVMap[swt.ID]))
	}
	return result, nil
}

func (s *SiteService) GetByID(id int64) (*models.SiteWithTags, error) {
	site, err := s.siteRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if site == nil {
		return nil, nil
	}

	tags, err := s.siteRepo.GetTags(site.ID)
	if err != nil {
		return nil, err
	}

	return &models.SiteWithTags{
		Site:        *site,
		Tags:        tags,
		DisplayTags: buildDisplayTags(tags),
		Color:       utils.GenerateColorFromName(site.Name),
		Initials:    utils.GetInitialsFromName(site.Name),
	}, nil
}

func (s *SiteService) GetByIDs(ids []int64) ([]models.SiteDisplay, error) {
	sitesWithTags, err := s.siteRepo.GetByIDsWithTags(ids)
	if err != nil {
		return nil, err
	}

	var result []models.SiteDisplay
	for _, swt := range sitesWithTags {
		result = append(result, buildSiteDisplay(swt, 0))
	}
	return result, nil
}

func (s *SiteService) Create(site *models.Site, tags []string) (int64, error) {
	id, err := s.siteRepo.Create(site)
	if err != nil {
		return 0, err
	}

	if len(tags) > 0 {
		if err := s.siteRepo.SetTags(id, tags); err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (s *SiteService) Update(site *models.Site, tags []string) error {
	if err := s.siteRepo.Update(site); err != nil {
		return err
	}

	if tags != nil {
		return s.siteRepo.SetTags(site.ID, tags)
	}
	return nil
}

func (s *SiteService) Delete(id int64) error {
	return s.siteRepo.Delete(id)
}

func (s *SiteService) GetCategories() ([]string, error) {
	return s.siteRepo.GetCategories()
}

func (s *SiteService) IncrementVisits(siteID int64, ip string, userID *int64) error {
	return s.siteRepo.IncrementVisits(siteID, ip, userID)
}

func (s *SiteService) GetSiteStats(siteID int64) (*models.SiteStats, error) {
	return s.siteRepo.GetSiteStats(siteID)
}

func (s *SiteService) GetAllSitesStats() ([]models.SiteStats, error) {
	return s.siteRepo.GetAllSitesStats()
}

func (s *SiteService) Search(query, category, sortBy string, page, pageSize int) ([]models.SiteDisplay, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	sites, total, err := s.siteRepo.SearchWithTags(query, category, sortBy, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	todayUVMap, err := s.siteRepo.GetAllSitesTodayUV()
	if err != nil {
		todayUVMap = make(map[int64]int64)
	}

	var result []models.SiteDisplay
	for _, swt := range sites {
		result = append(result, buildSiteDisplay(swt, todayUVMap[swt.ID]))
	}
	return result, total, nil
}

func (s *SiteService) GetSearchSuggestions(query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 5
	}
	return s.siteRepo.GetSearchSuggestions(query, limit)
}

func (s *SiteService) GetDashboardStats() (*models.DashboardStats, error) {
	return s.siteRepo.GetDashboardStats()
}

func (s *SiteService) GetTopSites(limit int) ([]models.Site, error) {
	if limit <= 0 {
		limit = 5
	}
	return s.siteRepo.GetTopSites(limit)
}

func (s *SiteService) GetCategoryStats() ([]models.CategoryStat, error) {
	return s.siteRepo.GetCategoryStats()
}
