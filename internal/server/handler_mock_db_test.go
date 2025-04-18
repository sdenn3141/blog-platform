package server_test

import (
	"blog-platform/internal/database"
	"blog-platform/internal/dto"
	"context"

	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) CreateBlog(ctx context.Context, create dto.BlogCreateDto) (*string, error) {
	args := m.Called(ctx, create)
	return args.Get(0).(*string), args.Error(1)
}

func (m *mockDB) GetBlogs(ctx context.Context) ([]*database.Blog, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*database.Blog), args.Error(1)
}

func (m *mockDB) GetBlog(ctx context.Context, id string) (*database.Blog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*database.Blog), args.Error(1)
}

func (m *mockDB) UpdateBlog(ctx context.Context, update dto.BlogUpdateDTO) (*database.Blog, error) {
	args := m.Called(ctx, update)
	return args.Get(0).(*database.Blog), args.Error(1)
}

func (m *mockDB) DeleteBlog(ctx context.Context, id string) (*database.Blog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*database.Blog), args.Error(1)
}

func (m *mockDB) GetBlogsByTerm(ctx context.Context, term string) ([]*database.Blog, error) {
	args := m.Called(ctx, term)
	return args.Get(0).([]*database.Blog), args.Error(1)
}

func (m *mockDB) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
