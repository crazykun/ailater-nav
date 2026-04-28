-- 添加全文索引
ALTER TABLE sites ADD FULLTEXT INDEX ft_sites_name_desc (name, description);
ALTER TABLE sites ADD FULLTEXT INDEX ft_sites_name (name);
ALTER TABLE tags ADD FULLTEXT INDEX ft_tags_name (name);
