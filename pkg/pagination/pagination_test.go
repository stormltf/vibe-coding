package pagination

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"
)

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     int
	}{
		{"first page", 1, 10, 0},
		{"second page", 2, 10, 10},
		{"third page", 3, 10, 20},
		{"custom page size", 2, 20, 20},
		{"large page", 100, 10, 990},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Page: tt.page, PageSize: tt.pageSize}
			if got := p.Offset(); got != tt.want {
				t.Errorf("Offset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPageResult(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		page      int
		pageSize  int
		wantPages int
	}{
		{"exact division", 100, 1, 10, 10},
		{"with remainder", 105, 1, 10, 11},
		{"single page", 5, 1, 10, 1},
		{"empty result", 0, 1, 10, 0},
		{"large dataset", 1000, 1, 20, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPageResult(nil, tt.total, tt.page, tt.pageSize)

			if result.Total != tt.total {
				t.Errorf("Total = %v, want %v", result.Total, tt.total)
			}
			if result.Page != tt.page {
				t.Errorf("Page = %v, want %v", result.Page, tt.page)
			}
			if result.PageSize != tt.pageSize {
				t.Errorf("PageSize = %v, want %v", result.PageSize, tt.pageSize)
			}
			if result.Pages != tt.wantPages {
				t.Errorf("Pages = %v, want %v", result.Pages, tt.wantPages)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if DefaultPage != 1 {
		t.Errorf("DefaultPage = %v, want 1", DefaultPage)
	}
	if DefaultPageSize != 10 {
		t.Errorf("DefaultPageSize = %v, want 10", DefaultPageSize)
	}
	if MaxPageSize != 100 {
		t.Errorf("MaxPageSize = %v, want 100", MaxPageSize)
	}
}

func newTestEngine() *route.Engine {
	opt := config.NewOptions([]config.Option{})
	return route.NewEngine(opt)
}

func TestGetFromQuery(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		wantPage     int
		wantPageSize int
	}{
		{"defaults", "", DefaultPage, DefaultPageSize},
		{"custom page", "page=5", 5, DefaultPageSize},
		{"custom page_size", "page_size=20", DefaultPage, 20},
		{"both custom", "page=3&page_size=25", 3, 25},
		{"exceeds max page_size", "page_size=200", DefaultPage, MaxPageSize},
		{"invalid page", "page=abc", DefaultPage, DefaultPageSize},
		{"negative page", "page=-1", DefaultPage, DefaultPageSize},
		{"zero page", "page=0", DefaultPage, DefaultPageSize},
		{"invalid page_size", "page_size=xyz", DefaultPage, DefaultPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestEngine()
			var gotPagination *Pagination

			r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
				gotPagination = GetFromQuery(c)
				c.String(http.StatusOK, "ok")
			})

			path := "/test"
			if tt.query != "" {
				path = "/test?" + tt.query
			}
			ut.PerformRequest(r, http.MethodGet, path, nil)

			if gotPagination.Page != tt.wantPage {
				t.Errorf("Page = %v, want %v", gotPagination.Page, tt.wantPage)
			}
			if gotPagination.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %v, want %v", gotPagination.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestNewPageResultWithData(t *testing.T) {
	data := []string{"a", "b", "c"}
	result := NewPageResult(data, 100, 2, 10)

	if result.List == nil {
		t.Error("List should not be nil")
	}
	if result.Total != 100 {
		t.Errorf("Total = %v, want 100", result.Total)
	}
	if result.Page != 2 {
		t.Errorf("Page = %v, want 2", result.Page)
	}
	if result.PageSize != 10 {
		t.Errorf("PageSize = %v, want 10", result.PageSize)
	}
	if result.Pages != 10 {
		t.Errorf("Pages = %v, want 10", result.Pages)
	}
}
