package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/porky256/rest-api/db"
	"github.com/porky256/rest-api/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// book represents data about a record book.

type Handler struct {
	Router   *gin.Engine
	DataBase db.Database
}

type ErrorMessage struct {
	Message string `json:"error"`
}

func InitializeHandler(database db.Database) Handler {
	handler := Handler{}
	handler.Router = gin.Default()
	handler.DataBase = database
	handler.Router.GET("/books", handler.getBooks)
	handler.Router.GET("/books/:id", handler.getBookByID)
	handler.Router.POST("/books", handler.postBook)
	handler.Router.DELETE("/books/:id", handler.deleteBook)
	handler.Router.PUT("/books/:id", handler.updateBook)
	return handler
}

// getBooks responds with the list of all books as JSON.
func (handler *Handler) getBooks(c *gin.Context) {
	filter := c.Request.URL.Query()
	if len(filter) != 0 {
		if !filter.Has("genre") {
			log.Println("invalid filter condition")
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid filter condition"})
			return
		}
		genre, err := strconv.Atoi(filter.Get("genre"))
		if err != nil || genre < 1 || genre > 3 {
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid filter condition"})
			return
		}
	}
	list, err := handler.DataBase.GetAllBooks(filter)

	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// postBook adds an book from JSON received in the request body.
func (handler *Handler) postBook(c *gin.Context) {
	var newBook models.Book

	// Call BindJSON to bind the received JSON to
	// newBook.

	if err := c.BindJSON(&newBook); err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid input"})
		return
	}

	// Add the new book to the slice.
	id, err := handler.DataBase.AddBook(newBook)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") { //unique_violation
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"input book name is not unique"})
			return
		}
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"internal server error"})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id})
}

func (handler *Handler) deleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid input"})
		return
	}

	err = handler.DataBase.DelBook(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusNotFound, ErrorMessage{"id not found"})
		default:
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"internal server error"})
		}
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (handler *Handler) updateBook(c *gin.Context) {
	var newBook models.Book

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid input"})
		return
	}

	if err := c.BindJSON(&newBook); err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid input"})
		return
	}

	err = handler.DataBase.UpdateBook(id, newBook)
	newBook.ID = id
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") { //unique_violation
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"input book name is not unique"})
			return
		}
		switch err {
		case sql.ErrNoRows:
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusNotFound, ErrorMessage{"id not found"})
		default:
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, newBook)
}

// getBookByID locates the book whose ID value matches the id
// parameter sent by the client, then returns that book as a response.
func (handler *Handler) getBookByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorMessage{"invalid input"})
		return
	}

	// Loop through the list of books, looking for
	// an book whose ID value matches the parameter.

	book, err := handler.DataBase.GetBookById(id)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusNotFound, ErrorMessage{"id not found"})
		default:
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorMessage{"internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, book)
}
