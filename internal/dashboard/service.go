package dashboard

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type DashboardService struct {
	repository *DashboardRepository
}

func NewDashboardService(repo *DashboardRepository) *DashboardService {
	return &DashboardService{
		repository: repo,
	}
}

func (s *DashboardService) GetDashboard(
	ctx context.Context,
	accountID string,
) (*DashboardStats, error) {
	var (
		account    Account
		totalSpent TotalSpent
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		account, err = s.repository.GetAccount(ctx, accountID)
		return err
	})

	g.Go(func() error {
		var err error
		totalSpent, err = s.repository.GetTotal(ctx, accountID)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &DashboardStats{
		Owner:      account.Owner,
		Balance:    account.Balance,
		Currency:   account.Currency,
		TotalSpent: totalSpent.Amount,
	}, nil
}

func (s *DashboardService) GetMonthlyStats(
	ctx context.Context,
	accountID string,
) (*MonthlyStats, error) {
	return s.repository.GetMonthlyStats(ctx, accountID)
}
