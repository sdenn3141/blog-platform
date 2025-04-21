package server

import (
	"blog-platform/internal/database"
	"blog-platform/internal/dto"
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) HealthHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()
	err := s.DB.Health(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"status": fmt.Sprintf("unhealthy %s", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) CreateBlogHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	blog := new(dto.BlogCreateDto)

	if err := c.Bind(blog); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	validate := validator.New()
	if err := validate.Struct(blog); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	createdId, err := s.DB.CreateBlog(ctx, *blog)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create blog %s", err)})
	}

	response := map[string]string{"data": *createdId}
	return c.JSON(http.StatusCreated, response)
}

func (s *Server) GetBlogHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()

	id := c.Param("id")
	data, err := s.DB.GetBlog(ctx, id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
	return c.JSON(http.StatusOK, data)
}

func (s *Server) GetBlogsHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()
	term := c.QueryParam("term")
	var data []*database.Blog
	var err error
	if term != "" {
		data, err = s.DB.GetBlogsByTerm(ctx, term)
	} else {
		data, err = s.DB.GetBlogs(ctx)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}

	return c.JSON(http.StatusOK, data)
}

func (s *Server) UpdateBlogHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()

	var updateBlog dto.BlogUpdateDTO
	updateBlog.Id = c.Param("id")

	if err := c.Bind(&updateBlog); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	validate := validator.New()
	if err := validate.Struct(updateBlog); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	data, err := s.DB.UpdateBlog(ctx, updateBlog)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
	return c.JSON(http.StatusOK, data)
}

func (s *Server) DeleteBlogHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 1*time.Second)
	defer cancel()
	id := c.Param("id")

	data, err := s.DB.DeleteBlog(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("internal server error - %v", err)})
	}
	return c.JSON(http.StatusOK, data)
}
