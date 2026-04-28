package repository

import (
	"ai-later-nav/internal/database"
	"ai-later-nav/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

type SiteRepository struct {
	db *sql.DB
}

func NewSiteRepository() *SiteRepository {
	return &SiteRepository{db: database.DB}
}

func (r *SiteRepository) CountSites() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM sites WHERE deleted = 0").Scan(&count)
	return count, err
}

func (r *SiteRepository) GetAll() ([]models.Site, error) {
	rows, err := r.db.Query(`
		SELECT id, name, url, description, logo, category, rating, visits, featured, deleted, created_at, updated_at
		FROM sites WHERE deleted = 0 ORDER BY featured DESC, rating DESC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query sites: %w", err)
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var s models.Site
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Description, &s.Logo,
			&s.Category, &s.Rating, &s.Visits, &s.Featured, &s.Deleted,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan site: %w", err)
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func (r *SiteRepository) GetByID(id int64) (*models.Site, error) {
	var s models.Site
	err := r.db.QueryRow(`
		SELECT id, name, url, description, logo, category, rating, visits, featured, deleted, created_at, updated_at
		FROM sites WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.URL, &s.Description, &s.Logo,
		&s.Category, &s.Rating, &s.Visits, &s.Featured, &s.Deleted,
		&s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query site by id: %w", err)
	}
	return &s, nil
}

func (r *SiteRepository) GetByName(name string) (*models.Site, error) {
	var s models.Site
	err := r.db.QueryRow(`
		SELECT id, name, url, description, logo, category, rating, visits, featured, deleted, created_at, updated_at
		FROM sites WHERE name = ? AND deleted = 0
	`, name).Scan(&s.ID, &s.Name, &s.URL, &s.Description, &s.Logo,
		&s.Category, &s.Rating, &s.Visits, &s.Featured, &s.Deleted,
		&s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query site by name: %w", err)
	}
	return &s, nil
}

func (r *SiteRepository) Create(site *models.Site) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO sites (name, url, description, logo, category, rating, featured)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, site.Name, site.URL, site.Description, site.Logo, site.Category, site.Rating, site.Featured)
	if err != nil {
		return 0, fmt.Errorf("insert site: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

func (r *SiteRepository) Update(site *models.Site) error {
	_, err := r.db.Exec(`
		UPDATE sites SET name=?, url=?, description=?, logo=?, category=?, rating=?, featured=?, updated_at=NOW()
		WHERE id=?
	`, site.Name, site.URL, site.Description, site.Logo, site.Category, site.Rating, site.Featured, site.ID)
	if err != nil {
		return fmt.Errorf("update site: %w", err)
	}
	return nil
}

func (r *SiteRepository) Delete(id int64) error {
	_, err := r.db.Exec("UPDATE sites SET deleted=1, updated_at=NOW() WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("soft delete site: %w", err)
	}
	return nil
}

func (r *SiteRepository) HardDelete(id int64) error {
	_, err := r.db.Exec("DELETE FROM sites WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("hard delete site: %w", err)
	}
	return nil
}

func (r *SiteRepository) Search(query, category, sortBy string, limit, offset int) ([]models.Site, int, error) {
	var args []interface{}
	where := []string{"s.deleted = 0"}

	if category != "" {
		where = append(where, "s.category = ?")
		args = append(args, category)
	}

	if query != "" {
		where = append(where, `(s.name LIKE ? OR s.description LIKE ? OR EXISTS (SELECT 1 FROM site_tags st JOIN tags t ON st.tag_id = t.id WHERE st.site_id = s.id AND t.name LIKE ?))`)
		args = append(args, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM sites s WHERE %s", whereClause)
	if err := r.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sites: %w", err)
	}

	orderBy := "s.featured DESC, s.rating DESC, s.name ASC"
	switch sortBy {
	case "name":
		orderBy = "s.name ASC"
	case "rating":
		orderBy = "s.rating DESC"
	case "visits":
		orderBy = "s.visits DESC"
	case "newest":
		orderBy = "s.created_at DESC"
	}

	selectArgs := append(args, limit, offset)
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT s.id, s.name, s.url, s.description, s.logo, s.category, s.rating, s.visits, s.featured, s.deleted, s.created_at, s.updated_at
		FROM sites s WHERE %s ORDER BY %s LIMIT ? OFFSET ?
	`, whereClause, orderBy), selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search sites: %w", err)
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var s models.Site
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Description, &s.Logo,
			&s.Category, &s.Rating, &s.Visits, &s.Featured, &s.Deleted,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan site: %w", err)
		}
		sites = append(sites, s)
	}
	return sites, total, rows.Err()
}

func (r *SiteRepository) GetSearchSuggestions(query string, limit int) ([]string, error) {
	if query == "" {
		return nil, nil
	}

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(`
		SELECT DISTINCT name
		FROM (
			SELECT name, rating, visits
			FROM sites
			WHERE deleted = 0 AND (name LIKE ? OR description LIKE ? OR id IN (
				SELECT st.site_id FROM site_tags st JOIN tags t ON st.tag_id = t.id
				WHERE t.name LIKE ?
			))
			ORDER BY rating DESC, visits DESC
		) AS ranked_sites
		LIMIT ?
	`, searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("query suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan suggestion: %w", err)
		}
		suggestions = append(suggestions, name)
	}
	return suggestions, rows.Err()
}

func (r *SiteRepository) GetCategories() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT category FROM sites WHERE deleted = 0 AND category != '' ORDER BY category")
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *SiteRepository) GetTags(siteID int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT t.name FROM tags t
		JOIN site_tags st ON t.id = st.tag_id
		WHERE st.site_id = ? ORDER BY t.name
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *SiteRepository) SetTags(siteID int64, tagNames []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM site_tags WHERE site_id = ?", siteID); err != nil {
		return fmt.Errorf("delete old tags: %w", err)
	}

	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var tagID int64
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", name).Scan(&tagID)
		if err == sql.ErrNoRows {
			result, err := tx.Exec("INSERT INTO tags (name) VALUES (?)", name)
			if err != nil {
				return fmt.Errorf("insert tag %s: %w", name, err)
			}
			tagID, _ = result.LastInsertId()
		} else if err != nil {
			return fmt.Errorf("query tag %s: %w", name, err)
		}

		if _, err := tx.Exec("INSERT IGNORE INTO site_tags (site_id, tag_id) VALUES (?, ?)", siteID, tagID); err != nil {
			return fmt.Errorf("link tag %s: %w", name, err)
		}
	}

	return tx.Commit()
}

func (r *SiteRepository) IncrementVisits(siteID int64, ip string, userID *int64) error {
	_, err := r.db.Exec("UPDATE sites SET visits = visits + 1 WHERE id = ?", siteID)
	if err != nil {
		return fmt.Errorf("increment visits: %w", err)
	}
	_, err = r.db.Exec("INSERT INTO visits (site_id, ip, user_id) VALUES (?, ?, ?)", siteID, ip, userID)
	return err
}

func (r *SiteRepository) GetSiteStats(siteID int64) (*models.SiteStats, error) {
	stats := &models.SiteStats{SiteID: siteID}

	err := r.db.QueryRow(`
		SELECT 
			COUNT(*) as pv,
			COUNT(DISTINCT ip) as uv
		FROM visits 
		WHERE site_id = ?
	`, siteID).Scan(&stats.PV, &stats.UV)
	if err != nil {
		return nil, fmt.Errorf("query total stats: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT 
			COUNT(*) as today_pv,
			COUNT(DISTINCT ip) as today_uv
		FROM visits 
		WHERE site_id = ? AND DATE(visited_at) = CURDATE()
	`, siteID).Scan(&stats.TodayPV, &stats.TodayUV)
	if err != nil {
		return nil, fmt.Errorf("query today stats: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT 
			COUNT(*) as week_pv,
			COUNT(DISTINCT ip) as week_uv
		FROM visits 
		WHERE site_id = ? AND visited_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
	`, siteID).Scan(&stats.WeekPV, &stats.WeekUV)
	if err != nil {
		return nil, fmt.Errorf("query week stats: %w", err)
	}

	return stats, nil
}

func (r *SiteRepository) GetAllSitesStats() ([]models.SiteStats, error) {
	rows, err := r.db.Query(`
		SELECT 
			site_id,
			COUNT(*) as pv,
			COUNT(DISTINCT ip) as uv,
			SUM(CASE WHEN DATE(visited_at) = CURDATE() THEN 1 ELSE 0 END) as today_pv,
			COUNT(DISTINCT CASE WHEN DATE(visited_at) = CURDATE() THEN ip END) as today_uv,
			SUM(CASE WHEN visited_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY) THEN 1 ELSE 0 END) as week_pv,
			COUNT(DISTINCT CASE WHEN visited_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY) THEN ip END) as week_uv
		FROM visits 
		GROUP BY site_id
	`)
	if err != nil {
		return nil, fmt.Errorf("query all stats: %w", err)
	}
	defer rows.Close()

	var stats []models.SiteStats
	for rows.Next() {
		var s models.SiteStats
		if err := rows.Scan(&s.SiteID, &s.PV, &s.UV, &s.TodayPV, &s.TodayUV, &s.WeekPV, &s.WeekUV); err != nil {
			return nil, fmt.Errorf("scan stats: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *SiteRepository) GetAllSitesTodayUV() (map[int64]int64, error) {
	rows, err := r.db.Query(`
		SELECT 
			site_id,
			COUNT(DISTINCT ip) as today_uv
		FROM visits 
		WHERE DATE(visited_at) = CURDATE()
		GROUP BY site_id
	`)
	if err != nil {
		return nil, fmt.Errorf("query today uv: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]int64)
	for rows.Next() {
		var siteID, todayUV int64
		if err := rows.Scan(&siteID, &todayUV); err != nil {
			return nil, fmt.Errorf("scan today uv: %w", err)
		}
		result[siteID] = todayUV
	}
	return result, nil
}

func (r *SiteRepository) GetAllWithTags() ([]models.SiteWithTags, error) {
	sites, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	var result []models.SiteWithTags
	for _, s := range sites {
		tags, err := r.GetTags(s.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, models.SiteWithTags{
			Site: s,
			Tags: tags,
		})
	}
	return result, nil
}

func (r *SiteRepository) SearchWithTags(query, category, sortBy string, limit, offset int) ([]models.SiteWithTags, int, error) {
	sites, total, err := r.Search(query, category, sortBy, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var result []models.SiteWithTags
	for _, s := range sites {
		tags, err := r.GetTags(s.ID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, models.SiteWithTags{
			Site: s,
			Tags: tags,
		})
	}
	return result, total, nil
}

func (r *SiteRepository) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// 总访问量
	err := r.db.QueryRow("SELECT COALESCE(SUM(visits), 0) FROM sites WHERE deleted = 0").Scan(&stats.TotalVisits)
	if err != nil {
		return nil, fmt.Errorf("query total visits: %w", err)
	}

	// 今日访问量
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM visits WHERE DATE(visited_at) = CURDATE()
	`).Scan(&stats.TodayVisits)
	if err != nil {
		return nil, fmt.Errorf("query today visits: %w", err)
	}

	// 本周访问量
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM visits WHERE visited_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
	`).Scan(&stats.WeekVisits)
	if err != nil {
		return nil, fmt.Errorf("query week visits: %w", err)
	}

	// 推荐站点数量
	err = r.db.QueryRow("SELECT COUNT(*) FROM sites WHERE deleted = 0 AND featured = 1").Scan(&stats.FeaturedCount)
	if err != nil {
		return nil, fmt.Errorf("query featured count: %w", err)
	}

	// 分类数量
	err = r.db.QueryRow("SELECT COUNT(DISTINCT category) FROM sites WHERE deleted = 0 AND category != ''").Scan(&stats.CategoryCount)
	if err != nil {
		return nil, fmt.Errorf("query category count: %w", err)
	}

	// 标签数量
	err = r.db.QueryRow("SELECT COUNT(*) FROM tags").Scan(&stats.TagCount)
	if err != nil {
		return nil, fmt.Errorf("query tag count: %w", err)
	}

	return stats, nil
}

func (r *SiteRepository) GetTopSites(limit int) ([]models.Site, error) {
	rows, err := r.db.Query(`
		SELECT id, name, url, description, logo, category, rating, visits, featured, deleted, created_at, updated_at
		FROM sites WHERE deleted = 0
		ORDER BY visits DESC, rating DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query top sites: %w", err)
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var s models.Site
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Description, &s.Logo,
			&s.Category, &s.Rating, &s.Visits, &s.Featured, &s.Deleted,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan site: %w", err)
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func (r *SiteRepository) GetCategoryStats() ([]models.CategoryStat, error) {
	rows, err := r.db.Query(`
		SELECT category, COUNT(*) as count
		FROM sites WHERE deleted = 0 AND category != ''
		GROUP BY category
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query category stats: %w", err)
	}
	defer rows.Close()

	var stats []models.CategoryStat
	for rows.Next() {
		var cs models.CategoryStat
		if err := rows.Scan(&cs.Category, &cs.Count); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		stats = append(stats, cs)
	}
	return stats, rows.Err()
}
