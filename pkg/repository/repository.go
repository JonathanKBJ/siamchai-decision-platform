package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ns.co.th/siamchai-decision-platform/pkg/models"
)

// MasterRepository handles Master Data CRUD & Upsert operations
type MasterRepository interface {
	UpsertBranches(ctx context.Context, branches []models.Branch) error
	GetBranchByID(ctx context.Context, id int) (*models.Branch, error)
	GetAllBranches(ctx context.Context) ([]models.Branch, error)

	UpsertProductBrands(ctx context.Context, brands []models.ProductBrand) error
	UpsertProductCategories(ctx context.Context, categories []models.ProductCategory) error
	UpsertProductGroups(ctx context.Context, groups []models.ProductGroup) error
	UpsertProductTypes(ctx context.Context, types []models.ProductType) error

	UpsertSuppliers(ctx context.Context, suppliers []models.Supplier) error
	UpsertCustomers(ctx context.Context, customers []models.Customer) error

	UpsertProducts(ctx context.Context, products []models.Product) error
	GetProductByID(ctx context.Context, id int) (*models.Product, error)
	GetAllProducts(ctx context.Context) ([]models.Product, error)
}

// SellRepository handles Sell Transaction operations
type SellRepository interface {
	UpsertSells(ctx context.Context, sells []models.Sell) error
	UpsertSellDetails(ctx context.Context, details []models.SellDetail) error
	GetSellsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Sell, error)
	GetSellDetailsBySellID(ctx context.Context, sellID int64) ([]models.SellDetail, error)
}

// StockRepository handles Stock Balance & Target operations
type StockRepository interface {
	UpsertStockBalance(ctx context.Context, items []models.StockBalance) error
	GetStockBalanceByShop(ctx context.Context, shopID int) ([]models.StockBalance, error)
	UpsertStockTargets(ctx context.Context, targets []models.StockTarget) error
	GetStockTargetsByShop(ctx context.Context, shopID int) ([]models.StockTarget, error)
}

// EmbeddingRepository handles Vector Embedding & Semantic Search
type EmbeddingRepository interface {
	SaveEmbedding(ctx context.Context, emb *models.ProductEmbedding) error
	GetEmbeddingByProductID(ctx context.Context, productID int) (*models.ProductEmbedding, error)
}

// GORM Implementation
type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *postgresRepository {
	return &postgresRepository{db: db}
}

// --- MasterRepository Implementation ---

func (r *postgresRepository) UpsertBranches(ctx context.Context, branches []models.Branch) error {
	if len(branches) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name", "is_active", "updated_at"}),
	}).Create(&branches).Error
}

func (r *postgresRepository) GetBranchByID(ctx context.Context, id int) (*models.Branch, error) {
	var b models.Branch
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *postgresRepository) GetAllBranches(ctx context.Context) ([]models.Branch, error) {
	var list []models.Branch
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *postgresRepository) UpsertProductBrands(ctx context.Context, brands []models.ProductBrand) error {
	if len(brands) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "status"}),
	}).Create(&brands).Error
}

func (r *postgresRepository) UpsertProductCategories(ctx context.Context, categories []models.ProductCategory) error {
	if len(categories) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "status"}),
	}).Create(&categories).Error
}

func (r *postgresRepository) UpsertProductGroups(ctx context.Context, groups []models.ProductGroup) error {
	if len(groups) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"description", "status"}),
	}).Create(&groups).Error
}

func (r *postgresRepository) UpsertProductTypes(ctx context.Context, types []models.ProductType) error {
	if len(types) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "status"}),
	}).Create(&types).Error
}

func (r *postgresRepository) UpsertSuppliers(ctx context.Context, suppliers []models.Supplier) error {
	if len(suppliers) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name", "tax_id", "address", "tel", "fax", "contact_name", "email", "status", "updated_at"}),
	}).Create(&suppliers).Error
}

func (r *postgresRepository) UpsertCustomers(ctx context.Context, customers []models.Customer) error {
	if len(customers) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name", "tax_id", "nation_id", "address", "tel", "email", "status", "updated_at"}),
	}).Create(&customers).Error
}

func (r *postgresRepository) UpsertProducts(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"pu_product_id", "brand_id", "category_id", "group_id", "type_id", "description", "status", "updated_at"}),
	}).Create(&products).Error
}

func (r *postgresRepository) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	var list []models.Product
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

// --- SellRepository Implementation ---

func (r *postgresRepository) UpsertSells(ctx context.Context, sells []models.Sell) error {
	if len(sells) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}, {Name: "sell_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"shop_id"}),
	}).Create(&sells).Error
}

func (r *postgresRepository) UpsertSellDetails(ctx context.Context, details []models.SellDetail) error {
	if len(details) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}, {Name: "sell_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"sell_id", "company_id", "product_id", "shop_id", "qty"}),
	}).Create(&details).Error
}

func (r *postgresRepository) GetSellsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Sell, error) {
	var list []models.Sell
	err := r.db.WithContext(ctx).Where("sell_date >= ? AND sell_date <= ?", startDate, endDate).Find(&list).Error
	return list, err
}

func (r *postgresRepository) GetSellDetailsBySellID(ctx context.Context, sellID int64) ([]models.SellDetail, error) {
	var list []models.SellDetail
	err := r.db.WithContext(ctx).Where("sell_id = ?", sellID).Find(&list).Error
	return list, err
}

// --- StockRepository Implementation ---

func (r *postgresRepository) UpsertStockBalance(ctx context.Context, items []models.StockBalance) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "shop_id"}, {Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"pu_product_id", "qty", "current_qty", "truesell_qty_7_day", "target_qty_7_day_original",
			"target_qty_7_day", "current_qty_by_type", "target_qty_by_type_7_day",
			"min_qty", "max_qty", "po_qty", "trans_in_qty", "trans_out_qty", "po_status",
			"truesell_qty_30_day", "target_qty_30_day_original", "target_qty_30_day",
			"target_qty_by_type_30_day", "updated_at",
		}),
	}).Create(&items).Error
}

func (r *postgresRepository) GetStockBalanceByShop(ctx context.Context, shopID int) ([]models.StockBalance, error) {
	var list []models.StockBalance
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&list).Error
	return list, err
}

func (r *postgresRepository) UpsertStockTargets(ctx context.Context, targets []models.StockTarget) error {
	if len(targets) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "shop_id"}, {Name: "product_group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"category_id", "type_id", "brand_id", "pu_product_id",
			"min_qty", "max_qty", "sell_multiply", "show_qty",
			"created_by", "updated_by", "updated_at",
		}),
	}).Create(&targets).Error
}

func (r *postgresRepository) GetStockTargetsByShop(ctx context.Context, shopID int) ([]models.StockTarget, error) {
	var list []models.StockTarget
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&list).Error
	return list, err
}

// --- EmbeddingRepository Implementation ---

func (r *postgresRepository) SaveEmbedding(ctx context.Context, emb *models.ProductEmbedding) error {
	if emb == nil {
		return fmt.Errorf("embedding cannot be nil")
	}
	return r.db.WithContext(ctx).Create(emb).Error
}

func (r *postgresRepository) GetEmbeddingByProductID(ctx context.Context, productID int) (*models.ProductEmbedding, error) {
	var emb models.ProductEmbedding
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&emb).Error
	if err != nil {
		return nil, err
	}
	return &emb, nil
}
