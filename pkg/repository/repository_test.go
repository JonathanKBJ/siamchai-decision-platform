package repository

import (
	"context"
	"testing"
	"time"

	"ns.co.th/siamchai-decision-platform/pkg/models"
)

func TestModelStructures(t *testing.T) {
	branch := models.Branch{
		ID:       1,
		Code:     "BKK01",
		Name:     "Bangkok Central Branch",
		IsActive: true,
	}

	if branch.ID != 1 {
		t.Errorf("Expected ID 1, got %d", branch.ID)
	}

	sell := models.Sell{
		ID:       1001,
		SellDate: time.Now(),
	}

	if sell.ID != 1001 {
		t.Errorf("Expected Sell ID 1001, got %d", sell.ID)
	}

	embedding := models.ProductEmbedding{
		Content:   "Sample product description embedding",
		Embedding: "[0.01, 0.02, 0.03]",
	}

	if embedding.Content == "" {
		t.Errorf("Expected non-empty content")
	}
}

type mockMasterRepo struct {
	branches map[int]models.Branch
}

func (m *mockMasterRepo) UpsertBranches(ctx context.Context, branches []models.Branch) error {
	for _, b := range branches {
		m.branches[b.ID] = b
	}
	return nil
}

func (m *mockMasterRepo) GetBranchByID(ctx context.Context, id int) (*models.Branch, error) {
	b, ok := m.branches[id]
	if !ok {
		return nil, nil
	}
	return &b, nil
}

func (m *mockMasterRepo) GetAllBranches(ctx context.Context) ([]models.Branch, error) {
	var res []models.Branch
	for _, b := range m.branches {
		res = append(res, b)
	}
	return res, nil
}

func (m *mockMasterRepo) UpsertProductBrands(ctx context.Context, brands []models.ProductBrand) error { return nil }
func (m *mockMasterRepo) UpsertProductCategories(ctx context.Context, categories []models.ProductCategory) error { return nil }
func (m *mockMasterRepo) UpsertProductGroups(ctx context.Context, groups []models.ProductGroup) error { return nil }
func (m *mockMasterRepo) UpsertProductTypes(ctx context.Context, types []models.ProductType) error { return nil }
func (m *mockMasterRepo) UpsertSuppliers(ctx context.Context, suppliers []models.Supplier) error { return nil }
func (m *mockMasterRepo) UpsertCustomers(ctx context.Context, customers []models.Customer) error { return nil }
func (m *mockMasterRepo) UpsertProducts(ctx context.Context, products []models.Product) error { return nil }
func (m *mockMasterRepo) GetProductByID(ctx context.Context, id int) (*models.Product, error) { return nil, nil }
func (m *mockMasterRepo) GetAllProducts(ctx context.Context) ([]models.Product, error) { return nil, nil }

func TestMasterRepositoryInterface(t *testing.T) {
	repo := &mockMasterRepo{branches: make(map[int]models.Branch)}
	ctx := context.Background()

	err := repo.UpsertBranches(ctx, []models.Branch{
		{ID: 1, Name: "Test Shop 1"},
		{ID: 2, Name: "Test Shop 2"},
	})
	if err != nil {
		t.Fatalf("Unexpected error upserting branches: %v", err)
	}

	b, err := repo.GetBranchByID(ctx, 1)
	if err != nil || b == nil {
		t.Fatalf("Expected SHOP 1 branch, got %v (err: %v)", b, err)
	}
	if b.Name != "Test Shop 1" {
		t.Errorf("Expected 'Test Shop 1', got '%s'", b.Name)
	}
}
