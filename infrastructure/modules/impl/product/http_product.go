package product

import (
	"bear/domain/entities"
	"bear/domain/usecases"
	"database/sql"
	"encoding/json"
	//setup "bear/infrastructure"
	au "bear/infrastructure/modules/impl/auth"
	"bear/infrastructure/modules/impl/http_error"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"io"
	"log"
	"net/http"
	"strconv"
)

type ProductModule struct {
	Db             *sql.DB
	Cookie         *securecookie.SecureCookie
	ProductUseCase usecases.ProductUseCase
}

func (c *ProductModule) Path() string {
	return "/product"
}

func (c *ProductModule) Setup(router *mux.Router) {
	//privateRoutes := router.PathPrefix("/private").Subrouter()
	//privateRoutes.Use(setup.AuthorizationMiddleware)

	privateRoutes := router.PathPrefix(c.Path()).Subrouter()

	privateRoutes.HandleFunc("/create", c.createProduct).Methods(http.MethodPost)
	privateRoutes.HandleFunc("/list", c.listProduct).Methods(http.MethodGet)
	privateRoutes.HandleFunc("/get/{productID}", c.getProductById).Methods(http.MethodGet)
	privateRoutes.HandleFunc("/update/{productID}", c.updateProduct).Methods(http.MethodPost)
	privateRoutes.HandleFunc("/delete/{productID}", c.deleteProduct).Methods(http.MethodDelete)
	privateRoutes.HandleFunc("/read-product", c.readProduct).Methods(http.MethodPost)
	privateRoutes.HandleFunc("/read-product/delete/{productID}", c.deleteReadProduct).Methods(http.MethodPost)
	privateRoutes.HandleFunc("/list/read-product", c.listReadProduct).Methods(http.MethodGet)
}

func (c *ProductModule) createProduct(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("[createProduct] Error ReadAll", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidRequestBody))
		return
	}

	var product entities.Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		log.Println("[createProduct] Error Unmarshal", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidRequestBody))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	ProductID, err := c.ProductUseCase.CreateProductUseCase(ctx, *user, product)
	if err != nil {
		log.Println("[createProduct] Error CreateProductUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(ProductID)
	if err != nil {
		log.Println("[createProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("[createProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) listProduct(w http.ResponseWriter, r *http.Request) {
	var filter entities.GeneralFilter
	var err error

	page := r.URL.Query().Get("page")
	if page != "" {
		filter.Page, err = strconv.ParseInt(page, 10, 64)
		if err != nil {
			log.Println("[listProduct] Error ParseInt page", err)
			http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
			return
		}
	}

	limit := r.URL.Query().Get("limit")
	if limit != "" {
		filter.Limit, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
			log.Println("[listProduct] Error ParseInt limit", err)
			return
		}
	}

	filter.Column = r.URL.Query().Get("orderBy")

	ordinationAsc := r.URL.Query().Get("ordinationAsc")
	if ordinationAsc == "true" {
		filter.OrdinationAsc = true
	}

	filter.Search = r.URL.Query().Get("search")

	if filter.Limit == 0 && filter.Page != 0 {
		log.Println("[listProduct] Error invalidParameter", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	response, err := c.ProductUseCase.ListProductUseCase(ctx, *user, filter)
	if err != nil {
		log.Println("[listProduct] Error ListProductUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Println("[listProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("[listProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) getProductById(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.Atoi(mux.Vars(r)["productID"])
	if err != nil {
		log.Println("[getProductById] Error Atoi ProductID", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	product, err := c.ProductUseCase.GetProductByIdUseCase(ctx, *user, int64(productID))
	if err != nil {
		log.Println("[getProductById] Error GetproductByIdUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(product)
	if err != nil {
		log.Println("[getProductById] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("[getProductById] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) updateProduct(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("[updateProduct] Error ReadAll", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidProduct))
		return
	}

	var product entities.Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		log.Println("[updateProduct] Error Unmarshal", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidProduct))
		return
	}

	productID, err := strconv.Atoi(mux.Vars(r)["productID"])
	if err != nil {
		log.Println("[updateProduct] Error Atoi ProductID", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
		return
	}

	product.Id = int64(productID)
	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)

	err = c.ProductUseCase.EditProductUseCase(ctx, *user, product)
	if err != nil {
		log.Println("[updateProduct] Error EditProductUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(entities.NewSuccessfulRequest())
	if err != nil {
		log.Println("[updateProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("[updateProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) deleteProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(mux.Vars(r)["productID"], 10, 64)
	if err != nil {
		log.Println("[deleteproduct] Error Atoi productID", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	err = c.ProductUseCase.DeleteProductUseCase(ctx, *user, productID)
	if err != nil {
		log.Println("[deleteProduct] Error DeleteProductUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(entities.NewSuccessfulRequest())
	if err != nil {
		log.Println("[deleteProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("[deleteProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) readProduct(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("[readProduct] Error ReadAll", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidRequestBody))
		return
	}

	var payload struct {
		Code string `json:"code"`
		Type string `json:"type"`
	}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Println("[readProduct] Error Unmarshal", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidRequestBody))
		return
	}

	if payload.Code == "" || payload.Type == "" {
		log.Println("[readProduct] Error Invalid Payload", payload)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
		return
	}

	var productID int64
	switch payload.Type {
	case "barcode":
		productID, err = strconv.ParseInt(payload.Code, 10, 64)
		if err != nil {
			log.Println("[readProduct] Error ParseInt for barcode", err)
			http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
			return
		}
	case "qrcode":
		productID, err = strconv.ParseInt(payload.Code, 10, 64)
		if err != nil {
			log.Println("[readProduct] Error ParseInt for qrcode", err)
			http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
			return
		}
	default:
		log.Println("[readProduct] Error Unsupported Type", payload.Type)
		http_error.HandleError(w, http_error.NewBadRequestError("Unsupported code type"))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	err = c.ProductUseCase.SetParamiter(ctx, *user, productID)
	if err != nil {
		log.Println("[readProduct] Error GetProductByIdUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(entities.NewSuccessfulRequest())
	if err != nil {
		log.Println("[readProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		log.Println("[readProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) deleteReadProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(mux.Vars(r)["productID"], 10, 64)
	if err != nil {
		log.Println("[DeleteReadProduct] Error Atoi productID", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	err = c.ProductUseCase.DeleteReadProduct(ctx, *user, productID)
	if err != nil {
		log.Println("[DeleteReadProduct] Error GetProductByIdUseCase", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(entities.NewSuccessfulRequest())
	if err != nil {
		log.Println("[DeleteReadProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		log.Println("[DeleteReadProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}

func (c *ProductModule) listReadProduct(w http.ResponseWriter, r *http.Request) {
	var filter entities.GeneralFilter
	var err error

	page := r.URL.Query().Get("page")
	if page != "" {
		filter.Page, err = strconv.ParseInt(page, 10, 64)
		if err != nil {
			log.Println("[listReadProduct] Error ParseInt page", err)
			http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
			return
		}
	}

	limit := r.URL.Query().Get("limit")
	if limit != "" {
		filter.Limit, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
			log.Println("[listReadProduct] Error ParseInt limit", err)
			return
		}
	}

	filter.Column = r.URL.Query().Get("orderBy")

	ordinationAsc := r.URL.Query().Get("ordinationAsc")
	if ordinationAsc == "true" {
		filter.OrdinationAsc = true
	}

	filter.Search = r.URL.Query().Get("search")

	if filter.Limit == 0 && filter.Page != 0 {
		log.Println("[listReadProduct] Error invalidParameter", err)
		http_error.HandleError(w, http_error.NewBadRequestError(http_error.InvalidParameter))
		return
	}

	ctx := r.Context()
	user := ctx.Value(au.CtxUserKey).(*entities.User)
	response, err := c.ProductUseCase.ListReadProduct(ctx, *user, filter)
	if err != nil {
		log.Println("[listReadProduct] Error ListReadProduct", err)
		http_error.HandleError(w, err)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Println("[listReadProduct] Error Marshal", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Println("[listReadProduct] Error Write", err)
		http_error.HandleError(w, http_error.NewUnexpectedError(http_error.Unexpected))
		return
	}
}
