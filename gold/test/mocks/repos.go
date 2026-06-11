package mocks

import (
	"sort"
	"sync"
	"time"

	"github.com/getsnowflake/snowflake/gold/src/models"
	"github.com/getsnowflake/snowflake/gold/src/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type MockAccountRepo struct {
	mu        sync.RWMutex
	accounts  map[string]*models.Account
	forUpdate map[string]chan struct{}
}

func NewMockAccountRepo() *MockAccountRepo {
	return &MockAccountRepo{
		accounts:  make(map[string]*models.Account),
		forUpdate: make(map[string]chan struct{}),
	}
}

func (m *MockAccountRepo) Seed(accounts ...*models.Account) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range accounts {
		m.accounts[a.ID.String()] = a
	}
}

func (m *MockAccountRepo) Lookup(id uuid.UUID) *models.Account {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accounts[id.String()]
}

func (m *MockAccountRepo) SetForUpdateBlock(id uuid.UUID, ch chan struct{}) {
	m.mu.Lock()
	m.forUpdate[id.String()] = ch
	m.mu.Unlock()
}

func (m *MockAccountRepo) FindByUserID(_ *gorm.DB, userID string) (*models.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.accounts {
		if a.UserID.String() == userID {
			cp := *a
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockAccountRepo) FindByID(_ *gorm.DB, id uuid.UUID) (*models.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.accounts[id.String()]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockAccountRepo) FindByIDForUpdate(_ *gorm.DB, id uuid.UUID) (*models.Account, error) {
	m.mu.RLock()
	ch, hasBlock := m.forUpdate[id.String()]
	m.mu.RUnlock()
	if hasBlock {
		<-ch
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.accounts[id.String()]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockAccountRepo) Create(_ *gorm.DB, account *models.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts[account.ID.String()] = account
	return nil
}

func (m *MockAccountRepo) Update(_ *gorm.DB, account *models.Account) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts[account.ID.String()] = account
	return nil
}

func (m *MockAccountRepo) All(_ *gorm.DB) ([]models.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]models.Account, 0, len(m.accounts))
	for _, a := range m.accounts {
		result = append(result, *a)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID.String() < result[j].ID.String()
	})
	return result, nil
}

func (m *MockAccountRepo) UpdateReconciliationFields(_ *gorm.DB, account *models.Account, fields map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[account.ID.String()]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, exists := fields["last_reconciled_at"]; exists {
		if t, ok := v.(time.Time); ok {
			a.LastReconciledAt = &t
		} else if v == nil {
			a.LastReconciledAt = nil
		}
	}
	if v, exists := fields["reconciliation_status"]; exists {
		if s, ok := v.(models.AccountReconciliationStatus); ok {
			a.ReconciliationStatus = s
		}
	}
	if v, exists := fields["discrepancy_detected_at"]; exists {
		if t, ok := v.(time.Time); ok {
			a.DiscrepancyDetectedAt = &t
		} else if v == nil {
			a.DiscrepancyDetectedAt = nil
		}
	}
	if v, exists := fields["discrepancy_amount"]; exists {
		if amount, ok := v.(int64); ok {
			a.DiscrepancyAmount = &amount
		} else if v == nil {
			a.DiscrepancyAmount = nil
		}
	}
	return nil
}

func (m *MockAccountRepo) DebitBalance(_ *gorm.DB, accountID uuid.UUID, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[accountID.String()]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	a.Balance -= amount
	return nil
}

func (m *MockAccountRepo) CreditBalance(_ *gorm.DB, accountID uuid.UUID, amount int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.accounts[accountID.String()]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	a.Balance += amount
	return nil
}

type MockTransactionRepo struct {
	mu           sync.RWMutex
	transactions map[string]*models.Transaction
}

func NewMockTransactionRepo() *MockTransactionRepo {
	return &MockTransactionRepo{transactions: make(map[string]*models.Transaction)}
}

func (m *MockTransactionRepo) Seed(transactions ...*models.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range transactions {
		m.transactions[t.ID.String()] = t
	}
}

func (m *MockTransactionRepo) All() []*models.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*models.Transaction, 0, len(m.transactions))
	for _, t := range m.transactions {
		result = append(result, t)
	}
	return result
}

func (m *MockTransactionRepo) FindByID(_ *gorm.DB, id uuid.UUID) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.transactions[id.String()]; ok {
		cp := *t
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockTransactionRepo) FindByIdempotencyKey(_ *gorm.DB, key uuid.UUID) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.transactions {
		if t.IdempotencyKey == key {
			cp := *t
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockTransactionRepo) FindByAccountID(_ *gorm.DB, accountID uuid.UUID, filter repository.TransactionFilter) ([]models.Transaction, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.Transaction
	for _, t := range m.transactions {
		if (t.SenderAccountID != nil && *t.SenderAccountID == accountID) ||
			(t.ReceiverAccountID != nil && *t.ReceiverAccountID == accountID) {
			if filter.Status != "" && string(t.Status) != filter.Status {
				continue
			}
			if filter.Type != "" && string(t.Type) != filter.Type {
				continue
			}
			result = append(result, *t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	total := int64(len(result))
	offset := (filter.Page - 1) * filter.Limit
	if offset >= len(result) {
		return []models.Transaction{}, total, nil
	}
	end := offset + filter.Limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *MockTransactionRepo) SumDailyAmount(_ *gorm.DB, accountID uuid.UUID, startOfDay time.Time) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var sum int64
	for _, t := range m.transactions {
		if t.SenderAccountID != nil && *t.SenderAccountID == accountID &&
			t.Status == models.TransactionStatusCompleted &&
			!t.CreatedAt.Before(startOfDay) {
			sum += t.Amount
		}
	}
	return sum, nil
}

func (m *MockTransactionRepo) CountDaily(_ *gorm.DB, accountID uuid.UUID, startOfDay time.Time) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, t := range m.transactions {
		if t.SenderAccountID != nil && *t.SenderAccountID == accountID &&
			t.Status == models.TransactionStatusCompleted &&
			!t.CreatedAt.Before(startOfDay) {
			count++
		}
	}
	return count, nil
}

func (m *MockTransactionRepo) Create(_ *gorm.DB, transaction *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.transactions {
		if t.IdempotencyKey == transaction.IdempotencyKey {
			return newUniqueViolation()
		}
	}
	m.transactions[transaction.ID.String()] = transaction
	return nil
}

func (m *MockTransactionRepo) Save(_ *gorm.DB, transaction *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions[transaction.ID.String()] = transaction
	return nil
}

type MockTransactionHistoryRepo struct {
	mu      sync.RWMutex
	entries map[string]*models.TransactionHistory
}

func NewMockTransactionHistoryRepo() *MockTransactionHistoryRepo {
	return &MockTransactionHistoryRepo{entries: make(map[string]*models.TransactionHistory)}
}

func (m *MockTransactionHistoryRepo) All() []models.TransactionHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]models.TransactionHistory, 0, len(m.entries))
	for _, e := range m.entries {
		result = append(result, *e)
	}
	return result
}

func (m *MockTransactionHistoryRepo) Create(_ *gorm.DB, history *models.TransactionHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	m.entries[history.ID.String()] = history
	return nil
}

func (m *MockTransactionHistoryRepo) FindByAccountID(_ *gorm.DB, accountID uuid.UUID, page, limit int) ([]models.TransactionHistory, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.TransactionHistory
	for _, e := range m.entries {
		if e.AccountID == accountID {
			result = append(result, *e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	total := int64(len(result))
	offset := (page - 1) * limit
	if offset >= len(result) {
		return []models.TransactionHistory{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *MockTransactionHistoryRepo) SumLedgerBalance(_ *gorm.DB, accountID uuid.UUID) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var balance int64
	for _, e := range m.entries {
		if e.AccountID == accountID {
			if e.EntryType == models.TransactionHistoryEntryTypeCredit {
				balance += e.Amount
			} else {
				balance -= e.Amount
			}
		}
	}
	return balance, nil
}

type MockFlakeRepo struct {
	mu     sync.RWMutex
	flakes map[string]*models.Flake
}

func NewMockFlakeRepo() *MockFlakeRepo {
	return &MockFlakeRepo{flakes: make(map[string]*models.Flake)}
}

func (m *MockFlakeRepo) Seed(flakes ...*models.Flake) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range flakes {
		m.flakes[f.ID.String()] = f
	}
}

func (m *MockFlakeRepo) All() []*models.Flake {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*models.Flake, 0, len(m.flakes))
	for _, f := range m.flakes {
		result = append(result, f)
	}
	return result
}

func (m *MockFlakeRepo) FindByAccountID(_ *gorm.DB, accountID uuid.UUID) ([]models.Flake, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.Flake
	for _, f := range m.flakes {
		if f.AccountID == accountID {
			result = append(result, *f)
		}
	}
	return result, nil
}

func (m *MockFlakeRepo) FindByKeyValue(_ *gorm.DB, keyValue string) (*models.Flake, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, f := range m.flakes {
		if f.KeyValue == keyValue && f.Status == models.FlakeStatusActive {
			cp := *f
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockFlakeRepo) FindByIDAndAccount(_ *gorm.DB, id uuid.UUID, accountID uuid.UUID) (*models.Flake, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if f, ok := m.flakes[id.String()]; ok && f.AccountID == accountID {
		cp := *f
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockFlakeRepo) FindByTypeAndAccount(_ *gorm.DB, keyType models.FlakeType, accountID uuid.UUID) (*models.Flake, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, f := range m.flakes {
		if f.KeyType == keyType && f.AccountID == accountID && f.Status == models.FlakeStatusActive {
			cp := *f
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockFlakeRepo) CountNonUnlimited(_ *gorm.DB, accountID uuid.UUID) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	notCounted := map[models.FlakeType]bool{
		models.FlakeTypeRandom: true,
		models.FlakeTypeHandle: true,
	}
	var count int64
	for _, f := range m.flakes {
		if f.AccountID == accountID && !notCounted[f.KeyType] && f.Status == models.FlakeStatusActive {
			count++
		}
	}
	return count, nil
}

func (m *MockFlakeRepo) Create(_ *gorm.DB, flake *models.Flake) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.flakes {
		if f.KeyValue == flake.KeyValue && f.Status == models.FlakeStatusActive {
			return newUniqueViolation()
		}
	}
	if flake.ID == uuid.Nil {
		flake.ID = uuid.New()
	}
	m.flakes[flake.ID.String()] = flake
	return nil
}

func (m *MockFlakeRepo) UpdateStatus(_ *gorm.DB, flake *models.Flake, status models.FlakeStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	flake.Status = status
	m.flakes[flake.ID.String()] = flake
	return nil
}

func newUniqueViolation() error {
	return &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}
}

type MockOutboxRepo struct {
	mu      sync.RWMutex
	entries map[string]*models.OutboxEvent
}

func NewMockOutboxRepo() *MockOutboxRepo {
	return &MockOutboxRepo{entries: make(map[string]*models.OutboxEvent)}
}

func (m *MockOutboxRepo) All() []models.OutboxEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]models.OutboxEvent, 0, len(m.entries))
	for _, e := range m.entries {
		result = append(result, *e)
	}
	return result
}

func (m *MockOutboxRepo) Create(_ *gorm.DB, entry *models.OutboxEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[entry.ID.String()] = entry
	return nil
}

func (m *MockOutboxRepo) ClaimPending(_ *gorm.DB, limit int) ([]models.OutboxEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.OutboxEvent
	for _, e := range m.entries {
		if e.Status == models.OutboxStatusPending {
			result = append(result, *e)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockOutboxRepo) MarkPublished(_ *gorm.DB, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[id.String()]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	e.Status = models.OutboxStatusPublished
	now := time.Now()
	e.PublishedAt = &now
	return nil
}

func (m *MockOutboxRepo) MarkFailed(_ *gorm.DB, id uuid.UUID, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[id.String()]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	e.Attempts++
	e.LastError = &errMsg
	if e.Attempts >= 5 {
		e.Status = models.OutboxStatusDead
	}
	return nil
}
