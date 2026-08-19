package models

import "time"

type Branch struct {
	ID        int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Code      string    `gorm:"column:code" json:"code"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Branch) TableName() string { return "branches" }

type ProductBrand struct {
	ID        int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Status    string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProductBrand) TableName() string { return "product_brands" }

type ProductCategory struct {
	ID        int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Status    string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProductCategory) TableName() string { return "product_categories" }

type ProductGroup struct {
	ID          int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Description string    `gorm:"column:description" json:"description"`
	Status      string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProductGroup) TableName() string { return "product_groups" }

type ProductType struct {
	ID        int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Status    string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProductType) TableName() string { return "product_types" }

type Product struct {
	ID          int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	PuProductID string    `gorm:"column:pu_product_id" json:"pu_product_id"`
	BrandID     *int      `gorm:"column:brand_id" json:"brand_id"`
	CategoryID  *int      `gorm:"column:category_id" json:"category_id"`
	GroupID     *int      `gorm:"column:group_id" json:"group_id"`
	TypeID      *int      `gorm:"column:type_id" json:"type_id"`
	Description string    `gorm:"column:description;not null" json:"description"`
	Status      string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Product) TableName() string { return "products" }

type Supplier struct {
	ID          int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Code        string    `gorm:"column:code" json:"code"`
	Name        string    `gorm:"column:name;not null" json:"name"`
	TaxID       string    `gorm:"column:tax_id" json:"tax_id"`
	Address     string    `gorm:"column:address" json:"address"`
	Tel         string    `gorm:"column:tel" json:"tel"`
	Fax         string    `gorm:"column:fax" json:"fax"`
	ContactName string    `gorm:"column:contact_name" json:"contact_name"`
	Email       string    `gorm:"column:email" json:"email"`
	Status      string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Supplier) TableName() string { return "suppliers" }

type Customer struct {
	ID        int       `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	Code      string    `gorm:"column:code" json:"code"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	TaxID     string    `gorm:"column:tax_id" json:"tax_id"`
	NationID  string    `gorm:"column:nation_id" json:"nation_id"`
	Address   string    `gorm:"column:address" json:"address"`
	Tel       string    `gorm:"column:tel" json:"tel"`
	Email     string    `gorm:"column:email" json:"email"`
	Status    string    `gorm:"column:status;default:'ACTIVE'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Customer) TableName() string { return "customers" }

type Sell struct {
	ID        int64     `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	ShopID    *int      `gorm:"column:shop_id" json:"shop_id"`
	SellDate  time.Time `gorm:"primaryKey;column:sell_date;not null" json:"sell_date"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Sell) TableName() string { return "sell" }

type SellDetail struct {
	ID        int64     `gorm:"primaryKey;column:id;autoIncrement:false" json:"id"`
	SellID    int64     `gorm:"column:sell_id;not null" json:"sell_id"`
	CompanyID *int      `gorm:"column:company_id" json:"company_id"`
	ProductID *int      `gorm:"column:product_id" json:"product_id"`
	ShopID    *int      `gorm:"column:shop_id" json:"shop_id"`
	Qty       int       `gorm:"column:qty;not null" json:"qty"`
	SellDate  time.Time `gorm:"primaryKey;column:sell_date;not null" json:"sell_date"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SellDetail) TableName() string { return "sell_detail" }

type StockBalance struct {
	ID                     int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID                 *int      `gorm:"column:shop_id;uniqueIndex:idx_stock_balance_shop_prod" json:"shop_id"`
	ProductID              *int      `gorm:"column:product_id;uniqueIndex:idx_stock_balance_shop_prod" json:"product_id"`
	PuProductID            *int      `gorm:"column:pu_product_id" json:"pu_product_id"`
	Qty                    int       `gorm:"column:qty;not null;default:0" json:"qty"`
	CurrentQty             float64   `gorm:"column:current_qty;default:0" json:"current_qty"`
	TruesellQty7Day        float64   `gorm:"column:truesell_qty_7_day;default:0" json:"truesell_qty_7_day"`
	TargetQty7DayOriginal  float64   `gorm:"column:target_qty_7_day_original;default:0" json:"target_qty_7_day_original"`
	TargetQty7Day          float64   `gorm:"column:target_qty_7_day;default:0" json:"target_qty_7_day"`
	CurrentQtyByType       float64   `gorm:"column:current_qty_by_type;default:0" json:"current_qty_by_type"`
	TargetQtyByType7Day    float64   `gorm:"column:target_qty_by_type_7_day;default:0" json:"target_qty_by_type_7_day"`
	MinQty                 float64   `gorm:"column:min_qty;default:0" json:"min_qty"`
	MaxQty                 float64   `gorm:"column:max_qty;default:0" json:"max_qty"`
	PoQty                  float64   `gorm:"column:po_qty;default:0" json:"po_qty"`
	TransInQty             float64   `gorm:"column:trans_in_qty;default:0" json:"trans_in_qty"`
	TransOutQty            float64   `gorm:"column:trans_out_qty;default:0" json:"trans_out_qty"`
	PoStatus               string    `gorm:"column:po_status;default:รอตัดสินใจ" json:"po_status"`
	TruesellQty30Day       float64   `gorm:"column:truesell_qty_30_day;default:0" json:"truesell_qty_30_day"`
	TargetQty30DayOriginal float64   `gorm:"column:target_qty_30_day_original;default:0" json:"target_qty_30_day_original"`
	TargetQty30Day         float64   `gorm:"column:target_qty_30_day;default:0" json:"target_qty_30_day"`
	TargetQtyByType30Day   float64   `gorm:"column:target_qty_by_type_30_day;default:0" json:"target_qty_by_type_30_day"`
	CreatedAt              time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StockBalance) TableName() string { return "stock_balance" }

type StockTarget struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID         *int      `gorm:"column:shop_id;uniqueIndex:idx_stock_targets_shop_group" json:"shop_id"`
	ProductGroupID *int      `gorm:"column:product_group_id;uniqueIndex:idx_stock_targets_shop_group" json:"product_group_id"`
	CategoryID     *int      `gorm:"column:category_id" json:"category_id"`
	TypeID         *int      `gorm:"column:type_id" json:"type_id"`
	BrandID        *int      `gorm:"column:brand_id" json:"brand_id"`
	PuProductID    string    `gorm:"column:pu_product_id" json:"pu_product_id"`
	MinQty         float64   `gorm:"column:min_qty;default:0" json:"min_qty"`
	MaxQty         float64   `gorm:"column:max_qty;default:0" json:"max_qty"`
	SellMultiply   float64   `gorm:"column:sell_multiply;default:0" json:"sell_multiply"`
	ShowQty        float64   `gorm:"column:show_qty;default:0" json:"show_qty"`
	CreatedBy      *int      `gorm:"column:created_by" json:"created_by"`
	UpdatedBy      *int      `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (StockTarget) TableName() string { return "stock_targets" }

type ProductEmbedding struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID *int      `gorm:"column:product_id" json:"product_id"`
	Content   string    `gorm:"column:content;not null" json:"content"`
	Embedding string    `gorm:"column:embedding;type:vector(1536)" json:"embedding"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (ProductEmbedding) TableName() string { return "product_embeddings" }
