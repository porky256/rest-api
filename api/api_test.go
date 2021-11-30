package api

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/porky256/rest-api/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestApiPostBook(t *testing.T) {
	type mockBehavior func(s *MockDatabase, book models.Book)
	gin.SetMode(gin.ReleaseMode)
	tests := []struct {
		name                 string
		inputBody            string
		inputBook            models.Book
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "Ok",
			inputBody: `{"name": "Book", "price": 0, "genre": 1, "amount": 0}`,
			inputBook: models.Book{
				Name:   "Book",
				Price:  0,
				Genre:  1,
				Amount: 0,
			},
			mockBehavior: func(r *MockDatabase, book models.Book) {
				r.EXPECT().AddBook(book).Return(1, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1}`,
		},
		//{
		//	name:                 "Name missing",
		//	inputBody:            `{"price": 1, "genre": 1, "amount": 5}`,
		//	mockBehavior:         func(r *mock_service.MockDatabase, book models.Book) {},
		//	expectedStatusCode:   http.StatusBadRequest,
		//	expectedResponseBody: `{"error":"invalid input"}`,
		//},
		{
			name:                 "Price missing",
			inputBody:            `{"name": error_book, "genre": 1, "amount": 5}`,
			mockBehavior:         func(r *MockDatabase, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},

		{
			name:                 "Genre missing",
			inputBody:            `{"name": error_book, "price": 1, "amount": 5}`,
			mockBehavior:         func(r *MockDatabase, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:                 "Amount missing",
			inputBody:            `{"name": error_book, "price": 1, "genre": 1}`,
			mockBehavior:         func(r *MockDatabase, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:                 "Invalid genre",
			inputBody:            `{"name": error_book, "price": 1, "genre": 0, "amount": 1}`,
			mockBehavior:         func(r *MockDatabase, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:                 "Invalid price",
			inputBody:            `{"name": error_book, "price": -1, "genre": 0, "amount": 1}`,
			mockBehavior:         func(r *MockDatabase, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:                 "Invalid amount",
			inputBody:            `{"name": error_book, "price": 1, "genre": 0, "amount": -1}`,
			mockBehavior:         func(r *MockDatabase, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:      "Ununique name",
			inputBody: `{"name": "repeated_test", "price": 1, "genre": 1, "amount": 1}`,
			inputBook: models.Book{
				Name:   "repeated_test",
				Price:  1,
				Genre:  1,
				Amount: 1,
			},
			mockBehavior: func(r *MockDatabase, book models.Book) {
				r.EXPECT().AddBook(book).Return(0, errors.New("duplicate key value"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"input book name is not unique"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			db := NewMockDatabase(c)
			test.mockBehavior(db, test.inputBook)
			//db := service.NewService(mockManager)
			rest_api := Handler{Router: gin.Default(), DataBase: db}

			r := gin.New()
			r.POST("/books", rest_api.postBook)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/books",
				bytes.NewBufferString(test.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestAPIGetByID(t *testing.T) {
	type mockBehavior func(s *MockDatabase, id interface{})
	gin.SetMode(gin.ReleaseMode)
	tests := []struct {
		name                 string
		inputID              interface{}
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:    "OK",
			inputID: 1,
			mockBehavior: func(r *MockDatabase, id interface{}) {
				r.EXPECT().GetBookById(id).Return(models.Book{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"name":"OK","price":1,"genre":1,"amount":1}`,
		},
		{
			name:                 "invalid id",
			inputID:              "invalid",
			mockBehavior:         func(r *MockDatabase, id interface{}) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:    "id not found",
			inputID: 256,
			mockBehavior: func(r *MockDatabase, id interface{}) {
				r.EXPECT().GetBookById(id).Return(models.Book{}, sql.ErrNoRows)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"id not found"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			db := NewMockDatabase(c)
			test.mockBehavior(db, test.inputID)
			//db := service.NewService(mockManager)
			rest_api := Handler{Router: gin.Default(), DataBase: db}

			r := gin.New()
			r.GET("/books/:id", rest_api.getBookByID)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/books/%v", test.inputID), nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestAPIDelByID(t *testing.T) {
	type mockBehavior func(s *MockDatabase, id interface{})
	gin.SetMode(gin.ReleaseMode)
	tests := []struct {
		name                 string
		inputID              interface{}
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:    "OK",
			inputID: 1,
			mockBehavior: func(r *MockDatabase, id interface{}) {
				r.EXPECT().DelBook(id).Return(nil)
			},
			expectedStatusCode:   http.StatusNoContent,
			expectedResponseBody: ``,
		},
		{
			name:                 "invalid id",
			inputID:              "invalid",
			mockBehavior:         func(r *MockDatabase, id interface{}) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:    "id not found",
			inputID: 256,
			mockBehavior: func(r *MockDatabase, id interface{}) {
				r.EXPECT().DelBook(id).Return(sql.ErrNoRows)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"id not found"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			db := NewMockDatabase(c)
			test.mockBehavior(db, test.inputID)
			//db := service.NewService(mockManager)
			rest_api := Handler{Router: gin.Default(), DataBase: db}

			r := gin.New()
			r.DELETE("/books/:id", rest_api.deleteBook)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/books/%v", test.inputID), nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestAPIUpdate(t *testing.T) {
	type mockBehavior func(s *MockDatabase, id interface{}, book models.Book)
	gin.SetMode(gin.ReleaseMode)
	tests := []struct {
		name                 string
		inputID              interface{}
		inputBook            models.Book
		inputBody            string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputID:   1,
			inputBook: models.Book{Name: "Updated", Price: 1, Genre: 1, Amount: 1},
			inputBody: `{"name":"Updated","price":1,"genre":1,"amount":1}`,
			mockBehavior: func(r *MockDatabase, id interface{}, book models.Book) {
				r.EXPECT().UpdateBook(id, book).Return(nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"name":"Updated","price":1,"genre":1,"amount":1}`,
		},
		{
			name:                 "invalid id",
			inputID:              "invalid",
			mockBehavior:         func(r *MockDatabase, id interface{}, book models.Book) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid input"}`,
		},
		{
			name:      "id not found",
			inputID:   256,
			inputBook: models.Book{Name: "Updated", Price: 1, Genre: 1, Amount: 1},
			inputBody: `{"name":"Updated","price":1,"genre":1,"amount":1}`,
			mockBehavior: func(r *MockDatabase, id interface{}, book models.Book) {
				r.EXPECT().UpdateBook(id, book).Return(sql.ErrNoRows)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"error":"id not found"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			db := NewMockDatabase(c)
			test.mockBehavior(db, test.inputID, test.inputBook)
			//db := service.NewService(mockManager)
			rest_api := Handler{Router: gin.Default(), DataBase: db}

			r := gin.New()
			r.PUT("/books/:id", rest_api.updateBook)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/books/%v", test.inputID),
				bytes.NewBufferString(test.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestAPIGet(t *testing.T) {
	type mockBehavior func(s *MockDatabase, filterCondition map[string][]string)
	gin.SetMode(gin.ReleaseMode)
	tests := []struct {
		name                 string
		filterCondition      map[string][]string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:            "OK",
			filterCondition: map[string][]string{},
			mockBehavior: func(r *MockDatabase, filterCondition map[string][]string) {
				r.EXPECT().GetAllBooks(filterCondition).Return([]models.Book{{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1},
					{ID: 1, Name: "OK2", Price: 1, Genre: 2, Amount: 1}}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `[{"id":1,"name":"OK","price":1,"genre":1,"amount":1},{"id":1,"name":"OK2","price":1,"genre":2,"amount":1}]`,
		},
		{
			name:            "OK with genre filter",
			filterCondition: map[string][]string{"genre": {"1"}},
			mockBehavior: func(r *MockDatabase, filterCondition map[string][]string) {
				r.EXPECT().GetAllBooks(filterCondition).Return([]models.Book{{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1}}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `[{"id":1,"name":"OK","price":1,"genre":1,"amount":1}]`,
		},
		{
			name:            "OK with name filter",
			filterCondition: map[string][]string{"name": {"OK"}},
			mockBehavior: func(r *MockDatabase, filterCondition map[string][]string) {
				r.EXPECT().GetAllBooks(filterCondition).Return([]models.Book{{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1}}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `[{"id":1,"name":"OK","price":1,"genre":1,"amount":1}]`,
		},
		{
			name:            "OK with both filter",
			filterCondition: map[string][]string{"name": {"OK"}, "genre": {"1"}},
			mockBehavior: func(r *MockDatabase, filterCondition map[string][]string) {
				r.EXPECT().GetAllBooks(filterCondition).Return([]models.Book{{ID: 1, Name: "OK", Price: 1, Genre: 1, Amount: 1}}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `[{"id":1,"name":"OK","price":1,"genre":1,"amount":1}]`,
		},
		{
			name:                 "invalid filter",
			filterCondition:      map[string][]string{"genre": {"some invalid info"}},
			mockBehavior:         func(r *MockDatabase, filterCondition map[string][]string) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid filter condition"}`,
		},
		{
			name:                 "another invalid filter",
			filterCondition:      map[string][]string{"some_filter_name": {"1"}},
			mockBehavior:         func(r *MockDatabase, filterCondition map[string][]string) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"error":"invalid filter condition"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			db := NewMockDatabase(c)
			test.mockBehavior(db, test.filterCondition)
			//db := service.NewService(mockManager)
			rest_api := Handler{Router: gin.Default(), DataBase: db}

			r := gin.New()
			r.GET("/books", rest_api.getBooks)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/books", nil)
			values := url.Values(test.filterCondition)
			req.URL.RawQuery = values.Encode()

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
