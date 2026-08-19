package ingestion

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"ns.co.th/siamchai-decision-platform/database"
	"ns.co.th/siamchai-decision-platform/pkg/models"
	"ns.co.th/siamchai-decision-platform/pkg/repository"
)

type DataMigrator struct {
	oraDB  *gorm.DB
	pgDB   *gorm.DB
	repo   repository.MasterRepository
	sRepo  repository.SellRepository
	stRepo repository.StockRepository
}

func NewDataMigrator(oraDB, pgDB *gorm.DB) *DataMigrator {
	repo := repository.NewPostgresRepository(pgDB)
	return &DataMigrator{
		oraDB:  oraDB,
		pgDB:   pgDB,
		repo:   repo,
		sRepo:  repo,
		stRepo: repo,
	}
}

// Oracle DTO structs (UPPERCASE column tags required for Oracle GORM driver)
type OraShop struct {
	ID     int    `gorm:"column:ID"`
	Code   string `gorm:"column:CODE"`
	Name   string `gorm:"column:NAME"`
	Active string `gorm:"column:ACTIVE"`
}

type OraBrand struct {
	ID     int    `gorm:"column:ID"`
	Name   string `gorm:"column:NAME"`
	Status string `gorm:"column:STATUS"`
}

type OraCategory struct {
	ID     int    `gorm:"column:ID"`
	Name   string `gorm:"column:NAME"`
	Status string `gorm:"column:STATUS"`
}

type OraGroup struct {
	ID          int    `gorm:"column:ID"`
	Description string `gorm:"column:DESCRIPTION"`
	Status      string `gorm:"column:STATUS"`
}

type OraType struct {
	ID     int    `gorm:"column:ID"`
	Name   string `gorm:"column:NAME"`
	Status string `gorm:"column:STATUS"`
}

type OraProduct struct {
	ID                int    `gorm:"column:ID"`
	PuProductID       string `gorm:"column:PU_PRODUCT_ID"`
	ProductBrandID    *int   `gorm:"column:PRODUCT_BRAND_ID"`
	ProductCategoryID *int   `gorm:"column:PRODUCT_CATEGORY_ID"`
	GroupID           *int   `gorm:"column:GROUP_ID"`
	ProductTypeID     *int   `gorm:"column:PRODUCT_TYPE_ID"`
	Description       string `gorm:"column:DESCRIPTION"`
	Status            string `gorm:"column:STATUS"`
}

type OraStockShop struct {
	ShopID    int `gorm:"column:SHOP_ID"`
	ProductID int `gorm:"column:PRODUCT_ID"`
	Qty       int `gorm:"column:QTY"`
}

type OraStockTarget struct {
	ID             int       `gorm:"column:ID"`
	ShopID         int       `gorm:"column:SHOP_ID"`
	ProductGroupID int       `gorm:"column:PRODUCT_GROUP_ID"`
	MinQty         float64   `gorm:"column:MIN_QTY"`
	MaxQty         float64   `gorm:"column:MAX_QTY"`
	SellMultiply   float64   `gorm:"column:SELL_MULTIPLY"`
	ShowQty        float64   `gorm:"column:SHOW_QTY"`
	CategoryID     *int      `gorm:"column:CATEGORY_ID"`
	TypeID         *int      `gorm:"column:TYPE_ID"`
	BrandID        *int      `gorm:"column:BRAND_ID"`
	CreatedBy      *int      `gorm:"column:CREATED_BY"`
	UpdatedBy      *int      `gorm:"column:UPDATED_BY"`
	CreatedAt      time.Time `gorm:"column:CREATED_AT"`
	UpdatedAt      time.Time `gorm:"column:UPDATED_AT"`
}

type OraCompany struct {
	ID       int    `gorm:"column:ID"`
	Type     string `gorm:"column:TYPE"`
	Code     string `gorm:"column:CODE"`
	Name     string `gorm:"column:NAME"`
	TaxID    string `gorm:"column:TAX_ID"`
	NationID string `gorm:"column:NATIONID"`
	Address  string `gorm:"column:ADDRESS"`
	Tel      string `gorm:"column:TEL"`
	Fax      string `gorm:"column:FAX"`
	Contact1 string `gorm:"column:CONTACT1"`
	Email1   string `gorm:"column:EMAIL1"`
}

type OraSell struct {
	ID       int64     `gorm:"column:ID"`
	ShopID   *int      `gorm:"column:SHOP_ID"`
	SellDate time.Time `gorm:"column:SELL_DATE"`
}

type OraSellDetail struct {
	ID        int64     `gorm:"column:ID"`
	SellID    int64     `gorm:"column:SELL_ID"`
	CompanyID *int      `gorm:"column:COMPANY_ID"`
	ProductID *int      `gorm:"column:PRODUCT_ID"`
	ShopID    *int      `gorm:"column:SHOP_ID"`
	Qty       int       `gorm:"column:QTY"`
	SellDate  time.Time `gorm:"column:SELL_DATE"`
}

// TruncateTargetTables clears existing PostgreSQL tables to ensure clean migration
func (m *DataMigrator) TruncateTargetTables(ctx context.Context) error {
	log.Println("[ETL Reset] Dropping and Recreating target PostgreSQL tables for clean migration...")
	sqlDrop := `DROP TABLE IF EXISTS sell_detail_default, sell_default, sell_detail, sell, sales_items, sales, stock_balance, stock_targets, product_embeddings, products, suppliers, customers, product_brands, product_categories, product_groups, product_types, branches CASCADE;`
	if err := m.pgDB.WithContext(ctx).Exec(sqlDrop).Error; err != nil {
		log.Printf("Warning dropping tables: %v", err)
	}

	sqlBytes, err := os.ReadFile("migrations/0001_init_schema.sql")
	if err != nil {
		sqlBytes, err = os.ReadFile("../migrations/0001_init_schema.sql")
	}
	if err == nil {
		if err := m.pgDB.WithContext(ctx).Exec(string(sqlBytes)).Error; err != nil {
			return fmt.Errorf("re-creating schema failed: %w", err)
		}
	}
	return nil
}

// RunMigration executes data migration with mode options
func (m *DataMigrator) RunMigration(ctx context.Context, mode, salesStartDate, salesEndDate string, fullReset bool) error {
	log.Println("==========================================================")
	log.Println(" Starting Data Migration Execution Engine")
	log.Println(" Mode:", mode)
	log.Println(" Sales Date Range Filter:", salesStartDate, "to", salesEndDate)
	log.Println(" Full Reset (Drop & Recreate Tables):", fullReset)
	log.Println("==========================================================")

	startTime := time.Now()

	// 0. Perform Full Reset ONLY if fullReset = true or mode = "full"
	if fullReset || mode == "full" {
		log.Println("[Reset Guard] Full reset requested. Executing table drop and schema recreation...")
		if err := m.TruncateTargetTables(ctx); err != nil {
			log.Printf("Warning during truncation: %v", err)
		}
	}

	// 1. Ensure Partition tables exist for requested range
	startT, err := time.Parse("2006-01-02", salesStartDate)
	if err != nil {
		startT = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	endT, err := time.Parse("2006-01-02", salesEndDate)
	if err != nil {
		endT = time.Now()
	}
	if err := database.EnsureMonthlyPartitionsForRange(m.pgDB, startT, endT); err != nil {
		log.Printf("Warning ensuring range partitions: %v", err)
	}

	// 2. Migrate Branches (ALL Shops per user instruction)
	if err := m.migrateBranches(ctx); err != nil {
		return fmt.Errorf("branch migration failed: %w", err)
	}

	// 2.5 Migrate Companies (Suppliers & Customers)
	if err := m.migrateCompanies(ctx); err != nil {
		return fmt.Errorf("company migration failed: %w", err)
	}

	// 3. Migrate Brands
	if err := m.migrateBrands(ctx); err != nil {
		return fmt.Errorf("brand migration failed: %w", err)
	}

	// 4. Migrate Categories
	if err := m.migrateCategories(ctx); err != nil {
		return fmt.Errorf("category migration failed: %w", err)
	}

	// 5. Migrate Groups
	if err := m.migrateGroups(ctx); err != nil {
		return fmt.Errorf("group migration failed: %w", err)
	}

	// 6. Migrate Types
	if err := m.migrateTypes(ctx); err != nil {
		return fmt.Errorf("type migration failed: %w", err)
	}

	// 7. Migrate Products
	if err := m.migrateProducts(ctx); err != nil {
		return fmt.Errorf("product migration failed: %w", err)
	}

	// 8. Migrate Stock Balance & Stock Targets
	if err := m.migrateStockBalance(ctx); err != nil {
		return fmt.Errorf("stock balance migration failed: %w", err)
	}
	if err := m.migrateStockTargets(ctx); err != nil {
		return fmt.Errorf("stock targets migration failed: %w", err)
	}

	// 9. Migrate Sell & Sell Detail Data
	if err := m.migrateSellData(ctx, salesStartDate, salesEndDate); err != nil {
		return fmt.Errorf("sell migration failed: %w", err)
	}

	log.Println("==========================================================")
	log.Printf(" Data Migration Completed Successfully in %v!", time.Since(startTime))
	log.Println("==========================================================")
	return nil
}

// RunFullMigration backward compatibility wrapper for full reset migration
func (m *DataMigrator) RunFullMigration(ctx context.Context, salesStartDate, salesEndDate string) error {
	return m.RunMigration(ctx, "full", salesStartDate, salesEndDate, true)
}

func (m *DataMigrator) migrateBranches(ctx context.Context) error {
	log.Println("[ETL 1/8] Migrating Branches (All Shops)...")
	var oraShops []OraShop
	if err := m.oraDB.Table("shop").Scan(&oraShops).Error; err != nil {
		return err
	}

	var pgBranches []models.Branch
	for _, s := range oraShops {
		pgBranches = append(pgBranches, models.Branch{
			ID:       s.ID,
			Code:     s.Code,
			Name:     s.Name,
			IsActive: s.Active == "Y",
		})
	}

	if err := m.repo.UpsertBranches(ctx, pgBranches); err != nil {
		return err
	}
	log.Printf(" -> Successfully migrated %d branches.", len(pgBranches))
	return nil
}

func (m *DataMigrator) migrateBrands(ctx context.Context) error {
	log.Println("[ETL 2/8] Migrating Product Brands...")
	var list []OraBrand
	if err := m.oraDB.Table("master_product_brand").Scan(&list).Error; err != nil {
		return err
	}

	var pgList []models.ProductBrand
	for _, item := range list {
		status := item.Status
		if status == "" {
			status = "ACTIVE"
		}
		pgList = append(pgList, models.ProductBrand{
			ID:     item.ID,
			Name:   item.Name,
			Status: status,
		})
	}
	if err := m.repo.UpsertProductBrands(ctx, pgList); err != nil {
		return err
	}
	log.Printf(" -> Successfully migrated %d brands.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateCategories(ctx context.Context) error {
	log.Println("[ETL 3/8] Migrating Product Categories...")
	var list []OraCategory
	if err := m.oraDB.Table("master_product_category").Scan(&list).Error; err != nil {
		return err
	}

	var pgList []models.ProductCategory
	for _, item := range list {
		status := item.Status
		if status == "" {
			status = "ACTIVE"
		}
		pgList = append(pgList, models.ProductCategory{
			ID:     item.ID,
			Name:   item.Name,
			Status: status,
		})
	}
	if err := m.repo.UpsertProductCategories(ctx, pgList); err != nil {
		return err
	}
	log.Printf(" -> Successfully migrated %d categories.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateGroups(ctx context.Context) error {
	log.Println("[ETL 4/8] Migrating Product Groups...")
	var list []OraGroup
	if err := m.oraDB.Table("product_group").Scan(&list).Error; err != nil {
		return err
	}

	var pgList []models.ProductGroup
	for _, item := range list {
		status := item.Status
		if status == "" {
			status = "ACTIVE"
		}
		pgList = append(pgList, models.ProductGroup{
			ID:          item.ID,
			Description: item.Description,
			Status:      status,
		})
	}
	if err := m.repo.UpsertProductGroups(ctx, pgList); err != nil {
		return err
	}
	log.Printf(" -> Successfully migrated %d product groups.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateTypes(ctx context.Context) error {
	log.Println("[ETL 5/8] Migrating Product Types...")
	var list []OraType
	if err := m.oraDB.Table("master_product_type").Scan(&list).Error; err != nil {
		return err
	}
	if len(list) > 0 {
		log.Printf("DEBUG master_product_type: fetched %d rows, sample: %+v", len(list), list[0])
	}

	var pgList []models.ProductType
	for _, item := range list {
		status := item.Status
		if status == "" {
			status = "ACTIVE"
		}
		pgList = append(pgList, models.ProductType{
			ID:     item.ID,
			Name:   item.Name,
			Status: status,
		})
	}
	if err := m.repo.UpsertProductTypes(ctx, pgList); err != nil {
		return err
	}
	log.Printf(" -> Successfully migrated %d product types.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateProducts(ctx context.Context) error {
	log.Println("[ETL 6/8] Migrating Products...")
	var list []OraProduct
	if err := m.oraDB.Table("product").Scan(&list).Error; err != nil {
		return err
	}

	var validBrands, validCats, validGroups, validTypes []int
	m.pgDB.WithContext(ctx).Model(&models.ProductBrand{}).Pluck("id", &validBrands)
	m.pgDB.WithContext(ctx).Model(&models.ProductCategory{}).Pluck("id", &validCats)
	m.pgDB.WithContext(ctx).Model(&models.ProductGroup{}).Pluck("id", &validGroups)
	m.pgDB.WithContext(ctx).Model(&models.ProductType{}).Pluck("id", &validTypes)

	brandMap := make(map[int]bool)
	for _, id := range validBrands { brandMap[id] = true }
	catMap := make(map[int]bool)
	for _, id := range validCats { catMap[id] = true }
	groupMap := make(map[int]bool)
	for _, id := range validGroups { groupMap[id] = true }
	typeMap := make(map[int]bool)
	for _, id := range validTypes { typeMap[id] = true }

	var pgList []models.Product
	for _, item := range list {
		status := item.Status
		if status == "" {
			status = "ACTIVE"
		}
		desc := item.Description
		if desc == "" {
			desc = "Unspecified Product"
		}

		var brandID, catID, groupID, typeID *int
		if item.ProductBrandID != nil && brandMap[*item.ProductBrandID] {
			brandID = item.ProductBrandID
		}
		if item.ProductCategoryID != nil && catMap[*item.ProductCategoryID] {
			catID = item.ProductCategoryID
		}
		if item.GroupID != nil && groupMap[*item.GroupID] {
			groupID = item.GroupID
		}
		if item.ProductTypeID != nil && typeMap[*item.ProductTypeID] {
			typeID = item.ProductTypeID
		}

		pgList = append(pgList, models.Product{
			ID:          item.ID,
			PuProductID: item.PuProductID,
			BrandID:     brandID,
			CategoryID:  catID,
			GroupID:     groupID,
			TypeID:      typeID,
			Description: desc,
			Status:      status,
		})
	}

	batchSize := 500
	for i := 0; i < len(pgList); i += batchSize {
		end := i + batchSize
		if end > len(pgList) {
			end = len(pgList)
		}
		if err := m.repo.UpsertProducts(ctx, pgList[i:end]); err != nil {
			return err
		}
	}
	log.Printf(" -> Successfully migrated %d products.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateStockBalance(ctx context.Context) error {
	log.Println("[ETL 7/8] Migrating Stock Balance (paged)...")

	var validShops, validProds []int
	m.pgDB.WithContext(ctx).Model(&models.Branch{}).Pluck("id", &validShops)
	m.pgDB.WithContext(ctx).Model(&models.Product{}).Pluck("id", &validProds)
	shopMap := make(map[int]bool)
	for _, id := range validShops { shopMap[id] = true }
	prodMap := make(map[int]bool)
	for _, id := range validProds { prodMap[id] = true }

	type stockKey struct {
		shopID    int
		productID int
	}
	aggregated := make(map[stockKey]int)

	chunkSize := 20000
	offset := 0

	for {
		var list []OraStockShop
		err := m.oraDB.Table("stock_shop").
			Select("SHOP_ID, PRODUCT_ID, QTY").
			Limit(chunkSize).Offset(offset).
			Find(&list).Error

		if err != nil {
			return fmt.Errorf("failed scanning stock_shop at offset %d: %w", offset, err)
		}
		if len(list) == 0 {
			break
		}

		for _, item := range list {
			if item.ShopID > 0 && item.ProductID > 0 && shopMap[item.ShopID] && prodMap[item.ProductID] {
				key := stockKey{shopID: item.ShopID, productID: item.ProductID}
				aggregated[key] += item.Qty
			}
		}

		offset += len(list)
		if len(list) < chunkSize {
			break
		}
	}

	var pgList []models.StockBalance
	for k, qty := range aggregated {
		sID := k.shopID
		pID := k.productID
		pgList = append(pgList, models.StockBalance{
			ShopID:    &sID,
			ProductID: &pID,
			Qty:       qty,
		})
	}

	batchSize := 500
	for i := 0; i < len(pgList); i += batchSize {
		end := i + batchSize
		if end > len(pgList) {
			end = len(pgList)
		}
		if err := m.stRepo.UpsertStockBalance(ctx, pgList[i:end]); err != nil {
			return fmt.Errorf("upserting stock balance batch failed: %w", err)
		}
	}
	log.Printf(" -> Successfully aggregated and migrated %d stock balance records.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateStockTargets(ctx context.Context) error {
	log.Println("[ETL 7.5/8] Migrating Stock Targets (bs_pg_setting) (paged)...")

	var validShops []int
	m.pgDB.WithContext(ctx).Model(&models.Branch{}).Pluck("id", &validShops)
	shopMap := make(map[int]bool)
	for _, id := range validShops {
		shopMap[id] = true
	}

	type targetKey struct {
		shopID         int
		productGroupID int
	}
	aggregated := make(map[targetKey]models.StockTarget)
	aggregatedMaxID := make(map[targetKey]int)

	chunkSize := 20000
	offset := 0

	for {
		var list []OraStockTarget
		err := m.oraDB.Table("bs_pg_setting").
			Select("ID, SHOP_ID, PRODUCT_GROUP_ID, MIN_QTY, MAX_QTY, SELL_MULTIPLY, SHOW_QTY, CATEGORY_ID, TYPE_ID, BRAND_ID, CREATED_BY, UPDATED_BY, CREATED_AT, UPDATED_AT").
			Limit(chunkSize).Offset(offset).
			Find(&list).Error

		if err != nil {
			return fmt.Errorf("failed scanning bs_pg_setting at offset %d: %w", offset, err)
		}
		if len(list) == 0 {
			break
		}

		for _, item := range list {
			if item.ShopID > 0 && item.ProductGroupID > 0 && shopMap[item.ShopID] {
				sID := item.ShopID
				pgID := item.ProductGroupID
				key := targetKey{shopID: sID, productGroupID: pgID}

				if maxID, exists := aggregatedMaxID[key]; !exists || item.ID > maxID {
					aggregatedMaxID[key] = item.ID
					aggregated[key] = models.StockTarget{
						ShopID:         &sID,
						ProductGroupID: &pgID,
						MinQty:         item.MinQty,
						MaxQty:         item.MaxQty,
						SellMultiply:   item.SellMultiply,
						ShowQty:        item.ShowQty,
						CategoryID:     item.CategoryID,
						TypeID:         item.TypeID,
						BrandID:        item.BrandID,
						CreatedBy:      item.CreatedBy,
						UpdatedBy:      item.UpdatedBy,
						CreatedAt:      item.CreatedAt,
						UpdatedAt:      item.UpdatedAt,
					}
				}
			}
		}

		offset += len(list)
		if len(list) < chunkSize {
			break
		}
	}

	var pgList []models.StockTarget
	for _, target := range aggregated {
		pgList = append(pgList, target)
	}

	batchSize := 500
	for i := 0; i < len(pgList); i += batchSize {
		end := i + batchSize
		if end > len(pgList) {
			end = len(pgList)
		}
		if err := m.stRepo.UpsertStockTargets(ctx, pgList[i:end]); err != nil {
			return fmt.Errorf("upserting stock targets batch failed: %w", err)
		}
	}
	log.Printf(" -> Successfully aggregated and migrated %d stock target records from bs_pg_setting.", len(pgList))
	return nil
}

func (m *DataMigrator) migrateCompanies(ctx context.Context) error {
	log.Println("[ETL 2.5/8] Migrating Companies (Suppliers: Type='S')...")
	var list []OraCompany
	if err := m.oraDB.Table("company").Where("type = 'S'").Find(&list).Error; err != nil {
		return fmt.Errorf("failed scanning supplier companies: %w", err)
	}

	var pgSuppliers []models.Supplier
	for _, item := range list {
		pgSuppliers = append(pgSuppliers, models.Supplier{
			ID:          item.ID,
			Code:        item.Code,
			Name:        item.Name,
			TaxID:       item.TaxID,
			Address:     item.Address,
			Tel:         item.Tel,
			Fax:         item.Fax,
			ContactName: item.Contact1,
			Email:       item.Email1,
			Status:      "ACTIVE",
		})
	}

	if err := m.repo.UpsertSuppliers(ctx, pgSuppliers); err != nil {
		return fmt.Errorf("failed upserting suppliers: %w", err)
	}
	log.Printf(" -> Successfully migrated %d Suppliers.", len(pgSuppliers))
	return nil
}

func (m *DataMigrator) migrateSellData(
	ctx context.Context,
	startDateStr, endDateStr string,
) error {
	log.Printf("[ETL 8/8] Migrating Sell & Sell Detail (%s to %s)...", startDateStr, endDateStr)

	var validShops, validProds, validSuppliers []int
	m.pgDB.WithContext(ctx).Model(&models.Branch{}).Pluck("id", &validShops)
	m.pgDB.WithContext(ctx).Model(&models.Product{}).Pluck("id", &validProds)
	m.pgDB.WithContext(ctx).Model(&models.Supplier{}).Pluck("id", &validSuppliers)

	shopMap := make(map[int]bool)
	for _, id := range validShops { shopMap[id] = true }
	prodMap := make(map[int]bool)
	for _, id := range validProds { prodMap[id] = true }
	supplierMap := make(map[int]bool)
	for _, id := range validSuppliers { supplierMap[id] = true }

	// 1. Fetch Sell headers
	querySell := fmt.Sprintf("sell_date >= DATE '%s' AND sell_date <= TIMESTAMP '%s 23:59:59'", startDateStr, endDateStr)
	var oraSells []OraSell
	if err := m.oraDB.Table("sell").Select("ID, SHOP_ID, SELL_DATE").Where(querySell).Find(&oraSells).Error; err != nil {
		return fmt.Errorf("failed fetching oracle sells: %w", err)
	}
	log.Printf(" -> Found %d Sell Header records in Oracle.", len(oraSells))

	var pgSells []models.Sell
	for _, s := range oraSells {
		var shopID *int
		if s.ShopID != nil && shopMap[*s.ShopID] {
			shopID = s.ShopID
		}
		pgSells = append(pgSells, models.Sell{
			ID:       s.ID,
			ShopID:   shopID,
			SellDate: s.SellDate,
		})
	}

	// Upsert sell headers in batches
	batchSize := 500
	for i := 0; i < len(pgSells); i += batchSize {
		end := i + batchSize
		if end > len(pgSells) {
			end = len(pgSells)
		}
		if err := m.sRepo.UpsertSells(ctx, pgSells[i:end]); err != nil {
			return fmt.Errorf("upserting sells batch failed: %w", err)
		}
	}
	log.Printf(" -> Upserted %d Sell Header records to PostgreSQL.", len(pgSells))

	// 2. Fetch Sell Detail details with join to sell header (s.sell_date)
	querySellDetail := fmt.Sprintf(`s.sell_date >= DATE '%s' AND s.sell_date <= TIMESTAMP '%s 23:59:59'`, startDateStr, endDateStr)

	var oraDetails []OraSellDetail
	if err := m.oraDB.Table("sell_detail sd").
		Select("sd.id AS ID, sd.sell_id AS SELL_ID, sd.company_id AS COMPANY_ID, sd.product_id AS PRODUCT_ID, sd.shop_id AS SHOP_ID, sd.qty AS QTY, s.sell_date AS SELL_DATE").
		Joins("INNER JOIN sell s ON sd.sell_id = s.id").
		Where(querySellDetail).
		Find(&oraDetails).Error; err != nil {
		return fmt.Errorf("failed fetching oracle sell_details: %w", err)
	}

	log.Printf(" -> Found %d Sell Detail records in Oracle.", len(oraDetails))

	var pgDetails []models.SellDetail
	for _, d := range oraDetails {
		var shopID, prodID, companyID *int
		if d.ShopID != nil && shopMap[*d.ShopID] {
			shopID = d.ShopID
		}
		if d.ProductID != nil && prodMap[*d.ProductID] {
			prodID = d.ProductID
		}
		if d.CompanyID != nil && supplierMap[*d.CompanyID] {
			companyID = d.CompanyID
		}

		pgDetails = append(pgDetails, models.SellDetail{
			ID:        d.ID,
			SellID:    d.SellID,
			CompanyID: companyID,
			ProductID: prodID,
			ShopID:    shopID,
			Qty:       d.Qty,
			SellDate:  d.SellDate,
		})
	}

	// Upsert sell details in batches of 500
	for i := 0; i < len(pgDetails); i += batchSize {
		end := i + batchSize
		if end > len(pgDetails) {
			end = len(pgDetails)
		}
		if err := m.sRepo.UpsertSellDetails(ctx, pgDetails[i:end]); err != nil {
			return fmt.Errorf("upserting sell details batch failed: %w", err)
		}
	}

	log.Printf(" -> Successfully migrated %d Sell Detail records to PostgreSQL.", len(pgDetails))
	return nil
}
