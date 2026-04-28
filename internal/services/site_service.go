package services

import (
	"ai-later-nav/internal/database/repository"
	"ai-later-nav/internal/models"
	"ai-later-nav/internal/utils"
)

type SiteService struct {
	siteRepo *repository.SiteRepository
}

func NewSiteService() *SiteService {
	return &SiteService{
		siteRepo: repository.NewSiteRepository(),
	}
}

func (s *SiteService) GetAll() ([]models.SiteDisplay, error) {
	sites, err := s.siteRepo.GetAllWithTags()
	if err != nil {
		return nil, err
	}

	var result []models.SiteDisplay
	for _, swt := range sites {
		result = append(result, models.SiteDisplay{
			Site:     swt.Site,
			Tags:     swt.Tags,
			Color:    utils.GenerateColorFromName(swt.Name),
			Initials: utils.GetInitialsFromName(swt.Name),
		})
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
		Site:     *site,
		Tags:     tags,
		Color:    utils.GenerateColorFromName(site.Name),
		Initials: utils.GetInitialsFromName(site.Name),
	}, nil
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

func (s *SiteService) IncrementVisits(siteID int64, ip string) error {
	return s.siteRepo.IncrementVisits(siteID, ip)
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

	var result []models.SiteDisplay
	for _, swt := range sites {
		result = append(result, models.SiteDisplay{
			Site:     swt.Site,
			Tags:     swt.Tags,
			Color:    utils.GenerateColorFromName(swt.Name),
			Initials: utils.GetInitialsFromName(swt.Name),
		})
	}
	return result, total, nil
}
