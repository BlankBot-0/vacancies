package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"vacancies/vacancies_fetcher/internal"
)

//var _ struct{} = (*Queries)(nil)

type (
	Repository interface {
		AddVacancies(ctx context.Context, arg AddVacanciesParams) error
		GetVacanciesByKeyWord(ctx context.Context, arg GetVacanciesByKeyWordParams) ([]GetVacanciesByKeyWordRow, error)
		WithTx(tx pgx.Tx) *Queries
	}
	txBuilder interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	}
	db interface {
		DBTX
		txBuilder
	}
)

type Deps struct {
	Repository Repository
	TxBuilder  db
}
type Repo struct {
	Deps
}

func NewRepo(deps Deps) *Repo {
	return &Repo{
		Deps: deps,
	}
}

func (r *Repo) AddVacancies(ctx context.Context, arg AddVacanciesParams) error {
	tx, err := r.TxBuilder.Begin(ctx)
	if err != nil {
		return err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)
	qtx := r.Repository.WithTx(tx)

	err = qtx.AddVacancies(ctx, arg)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repo) GetVacanciesByKeyWord(ctx context.Context, keyWord string) ([]internal.VacancyDTO, error) {
	tx, err := r.TxBuilder.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)
	qtx := r.Repository.WithTx(tx)

	vacancies, err := qtx.GetVacanciesByKeyWord(ctx, GetVacanciesByKeyWordParams{
		KeyWord:   keyWord,
		OffsetVal: 0,
		LimitVal:  10,
	})
	if err != nil {
		return nil, err
	}

	vacanciesDTOs := make([]internal.VacancyDTO, len(vacancies))
	for i, vacancy := range vacancies {
		vacanciesDTOs[i] = internal.VacancyDTO{
			Href:  vacancy.Href,
			Title: vacancy.Title,
		}
	}
	return vacanciesDTOs, nil
}
