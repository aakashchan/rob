package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"rob/lib/aws"
	c "rob/lib/common/constants"
	"rob/lib/common/httperr"
	"rob/lib/common/httpsucc"
	"rob/lib/common/types"
	"rob/lib/data"
	"rob/lib/datastore"
	"rob/lib/feed"
	mw "rob/lib/middleware"
	payment "rob/lib/payment"
	"rob/lib/validate"
	//"rob/lib/queue"
	"rob/lib/session"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:loginHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	phone, password, err := validate.Login(r.FormValue(c.Phone), r.FormValue(c.Password))

	if err != nil {
		httperr.E(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	u, err := datastore.GetUserByPhone(phone)
	if err != nil {
		httperr.DB(w, "Failed to get user Details", &err)
		return
	}

	if u.Verified == 0 {
		httperr.E(w, http.StatusBadRequest, "User not verified", nil)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		if err.Error() == "ErrMismatchedHashAndPassword" {
			httperr.E(w, http.StatusBadRequest, "Wrong Password", nil)
			return
		}
		httperr.E(w, http.StatusInternalServerError, "Password Check failed", &err)
		return
	}

	j, err := json.Marshal(u)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal response", &err)
		return
	}

	sess := session.Instance(r)
	roleId, err := datastore.GetRole(u.Id)
	log.Debugf("Login roleId = %v", *roleId)
	if err != nil {
		httperr.DB(w, "Failed to get user Details", &err)
		return
	}

	sess.Values[c.Id] = u.Id
	sess.Values[c.Phone] = u.Phone.String
	sess.Values[c.RoleId] = *roleId
	sess.Save(r, w)
	w.Write(j)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:logoutHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	session.Empty(sess)
	sess.Save(r, w)
	httpsucc.SuccWithMessage(w, "SuccessFully Logged Out")

}

func updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:updateProfileHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var user types.User
	sess := session.Instance(r)
	user.FirstName.String = r.FormValue(c.FirstName)
	user.LastName.String = r.FormValue(c.LastName)
	user.Gender.String = r.FormValue(c.Gender)
	user.Id = sess.Values[c.Id].(int)

	if user.FirstName.String == "" {
		httperr.E(w, http.StatusBadRequest, "Name cannot be empty", nil)
		return
	}

	err := datastore.UpdateUserProfile(user)

	if err != nil {
		httperr.DB(w, "Failed to get user Details", &err)
		return
	}

	httpsucc.SuccWithMessage(w, "Updated SuccessFully!")
}

func initiateSignUpHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:initiateSignUpHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	phone, err := validate.InitiateSignUp(r.FormValue(c.Phone))

	if err != nil {
		httperr.E(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var u *types.User

	u, err = datastore.GetUserByPhone(phone)

	if err == sql.ErrNoRows {
		// User does not exist
		u = &types.User{}
		u.Phone = sql.NullString{String: phone, Valid: true}
		u.Code = strconv.Itoa(rand.Intn(9000) + 1000)
		u.Verified = 0

		err1 := datastore.AddUser(*u)
		if err1 != nil {
			httperr.DB(w, "Unable to Add user", &err)
			return
		}

	} else if err != nil {
		// Unexpected error
		httperr.E(w, http.StatusInternalServerError, err.Error(), nil)
		return
	} else {
		// No error
		if u.Verified == 1 {
			// Verified user
			httperr.E(w, http.StatusBadRequest, "User with same phone number exists", nil)
			return
		}
	}

	err = aws.SendOtp(u.Phone.String, u.Code)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to send sms", &err)
		return
	}

	var res types.InitiateSignUpResponse

	d0 := int(u.Code[0] - '0')
	d1 := int(u.Code[1] - '0')
	d2 := int(u.Code[2] - '0')
	d3 := int(u.Code[3] - '0')

	// hash that is incomplete (on purpose) yet good enough for partialValidation
	res.Sum = d0 + d1 + d2 + d3
	res.Product = (d0 + 1) * (d1 + 1) * (d2 + 1) * (d3 + 1)

	j, err := json.Marshal(res)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	w.Write(j)
}

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:signUpHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	phone, code, name, password, err := validate.SignUp(r.FormValue(c.Phone),
		r.FormValue(c.Code), r.FormValue(c.Name), r.FormValue(c.Password))

	if err != nil {
		httperr.E(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	passhash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Error processing the password", &err)
		return
	}

	u, err := datastore.GetUserByPhone(phone)

	log.Debugf("GetUserByPhone result=%+v, error=%v", u, err)

	if err == sql.ErrNoRows {
		// no user found
		httperr.E(w, http.StatusBadRequest, "Partial User not found", nil)
		return
	} else if err != nil {
		// unexpected error
		httperr.E(w, http.StatusInternalServerError, err.Error(), nil)
		return
	} else {
		// User found
		if u.Verified == 1 {
			// User cannot be verified already
			httperr.E(w, http.StatusBadRequest, "User already exists with this phone number", nil)
			return
		}
	}

	if u.Code != code {
		httperr.E(w, http.StatusBadRequest, "Invalid OTP", nil)
		return
	}

	u.Password = string(passhash)
	u.FirstName = sql.NullString{String: name, Valid: true}
	u.Verified = 1

	err = datastore.UpdateUserDetails(*u)
	if err != nil {
		httperr.DB(w, "Unable to Add User", &err)
		return
	}
	err = datastore.InsertRole(u.Id, c.UserRole)
	if err != nil {
		httperr.DB(w, "Unable to Add User", &err)
		return
	}

	sess := session.Instance(r)
	sess.Values[c.Id] = u.Id
	sess.Values[c.Phone] = u.Phone.String
	sess.Values[c.RoleId] = c.UserRole
	fmt.Println(sess)
	sess.Save(r, w)

	j, err := json.Marshal(u)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal response", &err)
		return
	}
	w.Write(j)
}

func getPostHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:getPostHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	pid := r.FormValue(c.PostId)

	post, err := data.GetPostMetaData(pid)
	if err != nil {
		httperr.DB(w, "Failed to fetch post from the DB", &err)
		return
	}

	posts := make([]types.Post, 1)
	posts[0] = *post

	j, err := json.Marshal(posts)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	w.Write(j)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:createPostHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var p types.Post

	if err := r.ParseForm(); err != nil {
		httperr.E(w, http.StatusBadRequest, fmt.Sprintf("Invalid Request Form"), &err)
		return
	}

	p, err := validate.CreatePost(r.FormValue(c.CardType),
		r.FormValue(c.Src), r.FormValue(c.DpSrc),
		r.FormValue(c.Title), r.FormValue(c.Description),
		r.FormValue(c.ButtonText), r.FormValue(c.Url),
		r.FormValue(c.Icon), r.FormValue(c.GradientStart),
		r.FormValue(c.GradientEnd), r.PostForm[c.ChildPosts])

	if err != nil {
		httperr.E(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	postId, err := datastore.AddPost(p)

	if err != nil {
		httperr.DB(w, "Failed to create a post", &err)
		return
	}
	log.Debugf("Created postId=%s", postId)

	j, err := json.Marshal(postId)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
	}
}

func createPostLinkHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:createPostLinkHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	postId := r.FormValue(c.PostId)
	mascotId := r.FormValue(c.MascotId)
	mId, err := strconv.Atoi(mascotId)
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Invalid Request Parameters", &err)
		return
	}

	// makse sure the post exists
	_, err = datastore.GetPostMetaData(postId)
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Invalid postId or post does not exists", &err)
		return
	}

	err = datastore.PostLink(mId, postId)
	if err != nil {
		httperr.DB(w, "Failed to create a post link", &err)
		return
	}
	//queue.AddItemToQueue(postId, mId)

	httpsucc.SuccWithMessage(w, "Post successfully linked to mascot")

}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:feedHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	ls, ms, f, err := validate.Feed(
		r.FormValue(c.LastSync),
		r.FormValue(c.MascotId),
		r.FormValue(c.Flag))

	if err != nil {
		httperr.E(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	feed, err := feed.Get(ls, ms, f)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to get Feed", &err)
		return
	}
	j, err := json.Marshal(feed)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
	}
}

func initiatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:initiatePaymentHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	orderId, err := validate.Payment(r.FormValue(c.OrderId))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	sess := session.Instance(r)
	// TODO(Anudeep): We do not work on email level anymore
	email := sess.Values[c.Email].(string)
	response, err := payment.InitiateTransaction(orderId, email)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to get Hash", &err)
		return
	}
	j, err := json.Marshal(&response)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
	}

}

func addProductHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:addProductHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var product types.Product
	var err error
	product.Id = bson.NewObjectId()
	product.Sku = r.FormValue(c.Sku)
	product.Title = r.FormValue(c.Title)
	product.Brand = r.FormValue(c.Brand)
	product.Quantity, err = strconv.Atoi(r.FormValue(c.Quantity))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Quantity not compatible", &err)
		return
	}
	product.Description = r.FormValue(c.Description)
	product.UnitPrice, err = strconv.Atoi(r.FormValue(c.UnitPrice))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Unit Price not compatible", &err)
		return
	}
	product.Summary = r.FormValue(c.Summary)
	product.Image = r.FormValue(c.Image)
	product.ThumbNail = r.FormValue(c.ThumbNail)
	product.Color = r.FormValue(c.Color)
	product.Size = r.FormValue(c.Size)

	err = datastore.AddProduct(product)

	if err != nil {
		httperr.DB(w, "Failed to create a Product ", &err)
		return
	}
	log.Debugf("Created productId=%s", product.Id)

	j, err := json.Marshal(product.Id)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:getProductHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	productId := r.FormValue("pId")

	product, err := datastore.GetProduct(productId)

	if err != nil {
		if err.Error() == "not found" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No Prouduct found with %s", productId), &err)
			return
		} else {
			httperr.DB(w, "Failed to retrieve the Product ", &err)
			return
		}
	}
	log.Debugf("Fetched product=%s", &product)

	j, err := json.Marshal(&product)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:getPostsHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	posts, err := datastore.GetPosts()

	if err != nil {
		httperr.DB(w, "Failed to retrieve the Posts ", &err)
		return
	}

	log.Debugf("Fetched posts=%s", &posts)

	j, err := json.Marshal(&posts)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func addSaleHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:addSaleHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var sale types.Sale
	var err error
	sale.Title = r.FormValue(c.Title)
	sale.Brand = r.FormValue(c.Brand)
	sale.ProductSku = r.FormValue(c.ProductSku)
	sale.Description = r.FormValue(c.Description)
	sale.ThumbNail = r.FormValue(c.ThumbNail)
	sale.StockUnits, err = strconv.Atoi(r.FormValue(c.StockUnits))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Stock value not compatible", &err)
		return
	}
	sale.SaleStartTime, err = strconv.ParseInt(r.FormValue(c.SaleStartTime), 10, 64)
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "SaleStartTime not compatible", &err)
		return
	}
	sale.SaleEndTime, err = strconv.ParseInt(r.FormValue(c.SaleEndTime), 10, 64)
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "SaleEndTime not compatible", &err)
		return
	}

	saleId, err := datastore.AddSale(sale)
	if err != nil {
		httperr.DB(w, "Failed to create a Sale Listing", &err)
		return
	}
	log.Debugf("Created sale=%s", saleId)
	saleIdString := strconv.Itoa(saleId) // to enable proper marshalling in tests
	j, err := json.Marshal(saleIdString)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func getSaleHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:getSaleHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	saleId, err := strconv.Atoi(r.FormValue("sId"))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Sale Id  not compatible", &err)
		return
	}

	sale, err := datastore.GetSale(saleId)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No sale exists for %d", saleId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Sale info ", &err)
		return
	}
	log.Debugf("Fetched sale=%v", &sale)

	j, err := json.Marshal(&sale)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func getSalesHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:getSalesHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sales, err := datastore.GetSales()

	if err != nil {
		httperr.DB(w, "Failed to retrieve the Sales info ", &err)
		return
	}

	log.Debugf("Fetched sales=%s", &sales)

	j, err := json.Marshal(&sales)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:createOrderHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var order types.Order
	var err error
	d, err := hex.DecodeString(r.FormValue(c.ProductId))
	if err != nil || len(d) != 12 {
		httperr.E(w, http.StatusBadRequest, "Invalid Product Id", &err)
		return
	}
	order.ProductId = r.FormValue(c.ProductId)
	exists, err := mw.CheckExistanceMongo(c.ProductCollection, order.ProductId)
	if !exists {
		httperr.E(w, http.StatusNotFound, fmt.Sprintf("No Product exists with ProductId: %s", order.ProductId), &err)
		return
	}
	inStock, err := datastore.IsProductInStock(order.ProductId)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Check product stock", &err)
		return
	}
	if !inStock {
		err = errors.New("Product out of stock")
		httperr.E(w, http.StatusNotFound, fmt.Sprintf("Product with id : %s , is no longer in Stock . Order was not placed", order.ProductId), &err)
		return
	}
	sess := session.Instance(r)
	order.UserId = sess.Values[c.Id].(int)

	order.OrderDate, err = strconv.ParseInt(r.FormValue(c.OrderDate), 10, 64)
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "OrderDate not compatible", &err)
		return
	}
	order.Price, err = strconv.Atoi(r.FormValue(c.Price))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Price not compatible", &err)
		return
	}
	order.Tax, err = strconv.Atoi(r.FormValue(c.Tax))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Tax not compatible", &err)
		return
	}

	order.ShippingCost, err = strconv.Atoi(r.FormValue(c.ShippingCost))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "ShippingCost not compatible", &err)
		return
	}

	exists, err = mw.CheckExistanceMysql(c.AddressTable, c.Id, r.FormValue(c.AddressId), true)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Coudnt not check id address exists", &err)
		return
	}
	if !exists {
		err = errors.New("Address Not Found")
		httperr.E(w, http.StatusNotFound, "Address not found with the given AddressId", &err)
		return
	}
	order.AddressId, err = strconv.Atoi(r.FormValue(c.AddressId))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "AddressId not compatible", &err)
		return
	}

	order.Amount, err = strconv.Atoi(r.FormValue(c.Amount))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Amount not compatible", &err)
		return
	}

	order.SaleId, err = strconv.Atoi(r.FormValue(c.SaleId))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "SaleId not compatible", &err)
		return
	}

	orderId, err := datastore.CreateOrder(order)
	if err != nil {
		httperr.DB(w, "Failed to create the order", &err)
		return
	}

	log.Debugf("Created order=%s", orderId)
	orderIdString := strconv.Itoa(orderId) // to enable proper marshalling in tests
	j, err := json.Marshal(orderIdString)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:getOrderHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	userid := sess.Values[c.Id].(int)
	roleid := sess.Values[c.RoleId].(int)

	orderId, err := strconv.Atoi(r.FormValue("oId"))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Order Id  not compatible", &err)
		return
	}

	order, err := datastore.GetOrder(orderId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No order exists for %d", orderId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Order info ", &err)
		return
	}
	if roleid != c.AdminRole && userid != order.UserId {
		httperr.E(w, http.StatusUnauthorized, "No such order belongs to the user", &err)
		return
	}

	log.Debugf("Fetched order=%v", &order)

	j, err := json.Marshal(&order)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

//Retreives userId from session and feteches all orders placed by the user
func getUserOrdersHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:myOrdersHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	id1 := sess.Values[c.Id].(int)
	orders, err := datastore.GetUserOrders(id1)
	if err != nil {
		httperr.DB(w, "Failed to retrieve your orders history ", &err)
		return
	}

	log.Debugf("Fetched orders=%+v", &orders)

	j, err := json.Marshal(&orders)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func feedbackHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:feedbackHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	phone := sess.Values[c.Phone].(string)
	typ := r.FormValue(c.Type)
	desc := r.FormValue(c.Description)

	if typ != c.Feedback && typ != c.ContactUs {
		httperr.E(w, http.StatusBadRequest, "Invalid type value", nil)
		return
	}

	if desc == "" {
		httperr.E(w, http.StatusBadRequest, "Description cannot be empty", nil)
		return
	}

	err := datastore.AddFeedback(phone, typ, desc)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to DB failed", &err)
		return
	}

	// Trigger an email to c.EmailInfo
	err = aws.SendEmail(c.EmailInfo, []string{c.EmailInfo},
		fmt.Sprintf("New %s request", typ),
		desc, desc)

	if err != nil {
		log.Error("Failed to trigger email on feedback request", err.Error())
		// Intentionally not sending httperr back to user
		// because as long as it's stored in db, we are okay with email not
		// being triggered
	}
}

func addAddressHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:addAdressHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var address types.Address
	var err error
	sess := session.Instance(r)
	address.UserId = sess.Values[c.Id].(int)
	address.Address = r.FormValue(c.Address)
	address.AddressType = r.FormValue(c.AddressType)
	address.City = r.FormValue(c.City)
	address.State = r.FormValue(c.State)
	address.PostalCode, err = strconv.Atoi(r.FormValue(c.PostalCode))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Postal Code not compatible", &err)
		return
	}
	address.Phone = r.FormValue(c.Phone)
	id, err := datastore.AddAddress(address)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to add the address", &err)
		return
	}
	log.Debugf("Added Address=%d", id)
	idString := strconv.Itoa(id) // to enable proper marshalling in tests
	j, err := json.Marshal(idString)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func editAddressHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:editAdressHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	roleid := sess.Values[c.RoleId].(int)
	userId := sess.Values[c.Id].(int)
	addressId, err := strconv.Atoi(r.FormValue(c.AddressId))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Address Id  not compatible", &err)
		return
	}

	address, err := datastore.GetAddress(addressId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No address exists for %d", addressId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Address info ", &err)
		return
	}
	if roleid != c.AdminRole && userId != address.UserId {
		httperr.E(w, http.StatusUnauthorized, "No such address belongs to the user", &err)
		return
	}

	address.Address = r.FormValue(c.Address)
	address.AddressType = r.FormValue(c.AddressType)
	address.City = r.FormValue(c.City)
	address.State = r.FormValue(c.State)
	address.PostalCode, err = strconv.Atoi(r.FormValue(c.PostalCode))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Postal Code not compatible", &err)
		return
	}
	address.Phone = r.FormValue(c.Phone)

	err = datastore.EditAddress(addressId, *address)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to edit the address", &err)
		return
	}
	httpsucc.SuccWithMessage(w, "Address Updated SuccessFully!")
	return

}

func getUserAddressHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:getAddressHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	id1 := sess.Values[c.Id].(int)
	addresses, err := datastore.GetUserAddresses(id1)
	if err != nil {
		httperr.DB(w, "Failed to retrieve your Saved addresses ", &err)
		return
	}

	log.Debugf("Fetched addresses=%+v", &addresses)

	j, err := json.Marshal(&addresses)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return
}

func getAddressHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:getAddressHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	sess := session.Instance(r)
	userid := sess.Values[c.Id].(int)
	roleid := sess.Values[c.RoleId].(int)

	addressId, err := strconv.Atoi(r.FormValue("aId"))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Address Id  not compatible", &err)
		return
	}

	address, err := datastore.GetAddress(addressId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No address exists for %d", addressId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Address info ", &err)
		return
	}
	if roleid != c.AdminRole && userid != address.UserId {
		httperr.E(w, http.StatusUnauthorized, "No such address belongs to the user", &err)
		return
	}

	log.Debugf("Fetched address=%v", &address)

	j, err := json.Marshal(&address)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func checkDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:checkDeliveryHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	_, err := strconv.Atoi(r.FormValue("pc"))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Postal Code  not compatible", &err)
		return
	}
	pc := r.FormValue("pc")
	var check bool
	check = datastore.CheckDelivery(pc)

	var result string
	if check {
		result = "Deliverable"
	} else {
		result = "Not Deliverable"
	}
	j, err := json.Marshal(&result)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

/*
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:paymentHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var trans types.Transaction
	var err error
	sess := session.Instance(r)
	trans.UserId = sess.Values[c.Id].(int)
	roleid := sess.Values[c.RoleId].(int)
	trans.Amount, err = strconv.Atoi(r.FormValue(c.Amount))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Amount not compatible ", &err)
		return
	}
	trans.OrderId, err = strconv.Atoi(r.FormValue(c.OrderId))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "OrderId not compatible ", &err)
		return
	}
	order, err := datastore.GetOrder(trans.OrderId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No order exists for %d", trans.OrderId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Order info ", &err)
		return
	}
	if roleid != c.AdminRole && trans.UserId != order.UserId {
		httperr.E(w, http.StatusUnauthorized, "No such order belongs to the user", &err)
		return
	}
	inStock, err := datastore.IsProductInStock(order.ProductId)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Check product stock", &err)
		return
	}
	if !inStock {
		err = errors.New("Product out of stock")
		httperr.E(w, http.StatusNotFound, fmt.Sprintf("Product with id : %s , is no longer in Stock . Order was not placed", order.ProductId), &err)
		return
	}
	err = datastore.DecrementStock(order.ProductId)
	if err != nil {
		log.Error("Could not decrement product quantity")
		httperr.E(w, http.StatusInternalServerError, "Failed to reserve product stock", &err)
		return
	}
	err = datastore.UpdateSaleStock(order.SaleId, -1)
	if err != nil {
		log.Error("Could not decrement sale stock")
		if err.Error() == "Product Out Of Stock" {
			httperr.E(w, http.StatusNotFound, "Product Out Of Stock", &err)
			return
		}
		httperr.E(w, http.StatusInternalServerError, "Failed to reserve product stock", &err)
		return
	}
	trans.TimeOfCreation = time.Now().UTC().UnixNano()
	res, err := datastore.PaymentGateway(trans)
	if err != nil {
		err1 := datastore.IncrementStock(order.ProductId)
		if err1 != nil {
			log.Error("Could not increment product quantity")
		}
		err2 := datastore.UpdateSaleStock(order.SaleId, 1)
		if err2 != nil {
			log.Error("Could not increment sale stock")

		}
		httperr.E(w, http.StatusInternalServerError, "Failed to process payment", &err)
		return
	}
	log.Debugf("Payment process for =%v", &res)
	j, err := json.Marshal(&res)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	err = datastore.UpdateOrderTransaction(res.TransId, res.PaymentStatus, trans.OrderId)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Could not update transaction Id in Order", &err)
		return
	}
	return

}
*/
func getStatusHandler(w http.ResponseWriter, r *http.Request) {

	var funcName = "main.go:getStatusHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	saleId, err := strconv.Atoi(r.FormValue("sId"))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "Sale Id  not compatible", &err)
		return
	}

	status, err := datastore.GetStatus(saleId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No sale exists for %d", saleId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Sale info ", &err)
		return
	}
	log.Debugf("Fetched status=%v", &status)

	j, err := json.Marshal(&status)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:placeOrderHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var ship types.Shipping
	var err error
	sess := session.Instance(r)
	ship.UserId = sess.Values[c.Id].(int)
	roleid := sess.Values[c.RoleId].(int)
	ship.OrderId, err = strconv.Atoi(r.FormValue(c.OrderId))
	if err != nil {
		httperr.E(w, http.StatusBadRequest, "OrderId not compatible", &err)
	}
	order, err := datastore.GetOrder(ship.OrderId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No order exists for %d", ship.OrderId), &err)
			return
		}
		httperr.DB(w, "Failed to retrieve the Order info ", &err)
		return
	}
	if roleid != c.AdminRole && ship.UserId != order.UserId {
		httperr.E(w, http.StatusUnauthorized, "No such order belongs to the user", &err)
		return
	}
	ship.AddressId = order.AddressId
	id, err := datastore.PlaceOrder(ship)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to place the order", &err)
		return
	}
	log.Debugf("Order placed =%d", id)
	idString := strconv.Itoa(id) // to enable proper marshalling in tests
	j, err := json.Marshal(idString)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal Response", &err)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Writing to the response failed", &err)
		return
	}
	return

}

func fbtokenHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:fbtokenHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var err error
	sess := session.Instance(r)
	userId := sess.Values[c.Id].(int)
	token := r.FormValue(c.Token)
	err = datastore.UpdateToken(userId, token)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to update the instance token", &err)
		return
	}
	httpsucc.SuccWithMessage(w, "Token Updated SuccessFully!")
	return

}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:resetPasswordHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	var err error
	phone := r.FormValue(c.Phone)
	token := r.FormValue(c.ResetPasswordToken)

	if !validate.ValidPhoneNumber(phone) {
		httperr.E(w, http.StatusBadRequest, "Invalid phone number.", nil)
		return
	}

	if len(token) < 4 {
		httperr.E(w, http.StatusBadRequest, "Invalid token number.", nil)
		return
	}

	u, err := datastore.GetUserByPhone(phone)
	if err != nil {
		if err == sql.ErrNoRows {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No User Exists for  %s", phone), &err)
			return
		}
		httperr.DB(w, "Failed to get user Details", &err)
		return
	}

	if u.Verified == 0 {
		httperr.E(w, http.StatusBadRequest, "User not verified", nil)
		return
	}

	if u.ResetPasswordToken.String != token {
		log.Error(fmt.Sprintf("want %s , got %s", u.ResetPasswordToken.String, token))
		err = errors.New("Incorrect Reset Token")
		httperr.E(w, http.StatusBadRequest, "Incorrect Reset Password Token Entered", &err)
		return
	}

	new1 := r.FormValue(c.NewPassword)
	if len(new1) < 8 {
		httperr.E(w, http.StatusBadRequest, "Password length needs to be >= 8", nil)
		return
	}

	new2 := r.FormValue(c.NewPasswordRepeat)
	if new1 != new2 {
		err = errors.New("New Password Mismatch")
		httperr.E(w, http.StatusBadRequest, "New Password Mismatch . Please make sure new password is entered correctly", &err)
		return
	}

	passhash, err := bcrypt.GenerateFromPassword([]byte(new1), bcrypt.DefaultCost)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Error processing the password", &err)
		return
	}

	newPassword := string(passhash)
	err = datastore.ResetPassword(phone, newPassword)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to reset password", &err)
		return
	}
	httpsucc.SuccWithMessage(w, "Password SuccessFully updated!")
}

func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:forgotPasswordHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	var err error
	phone := r.FormValue(c.Phone)

	if !validate.ValidPhoneNumber(phone) {
		httperr.E(w, http.StatusBadRequest, "Invalid phone number.", nil)
		return
	}

	u, err := datastore.GetUserByPhone(phone)
	if err != nil {
		if err == sql.ErrNoRows {
			httperr.E(w, http.StatusNotFound, fmt.Sprintf("No User Exists for  %s", phone), &err)
			return
		}
		httperr.DB(w, "Failed to get user Details", &err)
		return
	}

	if u.Verified == 0 {
		httperr.E(w, http.StatusBadRequest, "User not verified", nil)
		return
	}

	code := strconv.Itoa(rand.Intn(9000) + 1000)
	err = datastore.UpdatePasswordResetToken(phone, code)
	if err != nil {
		log.Error(err)
		httperr.E(w, http.StatusInternalServerError, "Failed to update password reset token", &err)
		return
	}

	err = aws.SendResetOtp(phone, code)
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to send reset code", &err)
		return
	}
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:cacheHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	// This is temporary function. This should be replaced later with the correct function

	var ss []string

	/*ss = append(ss, "https://cdn.twiq.in/r/tos.html")
	ss = append(ss, "https://cdn.twiq.in/r/privacy.html")
	ss = append(ss, "https://cdn.twiq.in/r/sales/demo-1.html")
	ss = append(ss, "https://cdn.twiq.in/img/dorothyperkins_store.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/dress_1.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/dress_4.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/take_money.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/demo-1.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/dress_3.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/review_1.png")
	ss = append(ss, "https://cdn.twiq.in/img/review_2.png")
	ss = append(ss, "https://cdn.twiq.in/img/btn_dull.png")
	ss = append(ss, "https://cdn.twiq.in/r/sales/demo-2.html")
	ss = append(ss, "https://cdn.twiq.in/img/demo-2.jpg")
	ss = append(ss, "https://cdn.twiq.in/img/demobrownie_2.jpg")
	*/

	ss, err := data.GetCacheDetails()
	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to retrieve from DB", &err)
	}

	j, err := json.Marshal(ss)

	if err != nil {
		httperr.E(w, http.StatusInternalServerError, "Failed to Marshal response", &err)
		return
	}

	w.Write(j)
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:urlHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	url := r.FormValue(c.Url)
	op := r.FormValue(c.Type)

	if op == "1" {
		err := data.InsertCacheUrl(url)
		if err != nil {
			httperr.E(w, http.StatusInternalServerError, "Couldnt insert url", &err)
			return
		}
	}

	if op == "0" {
		err := data.DeleteCacheUrl(url)
		if err != nil {
			httperr.E(w, http.StatusInternalServerError, "Couldnt delete url", &err)
			return
		}
	}

	httpsucc.SuccWithMessage(w, "Successfully completed operation")
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	var funcName = "main.go:deletePostHandler"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	postId := r.FormValue(c.PostId)
	err := datastore.DeletePost(postId)
	if err != nil {
		if err.Error() == "Invalid postId" {
			httperr.E(w, http.StatusBadRequest, "Invalid postId", &err)
			return
		}
		httperr.E(w, http.StatusInternalServerError, "Couldnt delete post", &err)
		return
	}
	httpsucc.SuccWithMessage(w, "Successfully Deleted Post ")
}

func vrHandler(w http.ResponseWriter, r *http.Request) {

	url := "twiq://verify.token?token=" + r.FormValue("token")
	http.Redirect(w, r, url, 302)
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	httpsucc.SuccWithMessage(w, "Server up and running")
}

func loginOkHandler(w http.ResponseWriter, r *http.Request) {
	httpsucc.SuccWithMessage(w, "Server up. Login ok")
}

func initLogging() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	if c.DE {
		log.SetLevel(log.DebugLevel)
	}
	//aws.DisableModule = true
	log.Infof("Log level set to: %s", log.GetLevel())
}

func getRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/ok", okHandler).Methods("GET", "POST")
	r.Handle("/login-ok",
		alice.New(mw.Auth).
			ThenFunc(loginOkHandler)).
		Methods("GET", "POST")

	r.Handle("/login",
		alice.New(mw.NoAuth).
			ThenFunc(loginHandler)).
		Methods("POST")

	r.Handle("/signup",
		alice.New(mw.NoAuth).
			ThenFunc(signUpHandler)).
		Methods("POST")

	r.Handle("/post",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole)).
			ThenFunc(createPostHandler)).
		Methods("POST")

	r.Handle("/post",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getPostHandler)).
		Methods("GET")

	r.Handle("/postlink",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole)).
			ThenFunc(createPostLinkHandler)).
		Methods("POST")
	r.Handle("/logout",
		alice.New(mw.Auth).
			ThenFunc(logoutHandler)).
		Methods("GET")

	r.Handle("/editprofile",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(updateProfileHandler)).
		Methods("POST")

	r.Handle("/feed",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(feedHandler)).
		Methods("POST")

	r.Handle("/product",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole)).
			ThenFunc(addProductHandler)).
		Methods("POST")

	r.Handle("/product",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getProductHandler)).
		Methods("GET")

	r.Handle("/sale",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole)).
			ThenFunc(addSaleHandler)).
		Methods("POST")

	r.Handle("/sale",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getSaleHandler)).
		Methods("GET")

	r.Handle("/sales",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getSalesHandler)).
		Methods("GET")

	r.Handle("/order",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(createOrderHandler)).
		Methods("POST")

	r.Handle("/order",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getOrderHandler)).
		Methods("GET")

	r.Handle("/orders",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getUserOrdersHandler)).
		Methods("GET")

	r.Handle("/address",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			Append(mw.ValidatePhone).
			ThenFunc(addAddressHandler)).
		Methods("POST")

	r.Handle("/editAddress",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			Append(mw.ValidatePhone).
			ThenFunc(editAddressHandler)).
		Methods("POST")

	r.Handle("/address",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getAddressHandler)).
		Methods("GET")

	r.Handle("/addresses",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getUserAddressHandler)).
		Methods("GET")

	r.Handle("/feedback",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(feedbackHandler)).
		Methods("POST")

	r.Handle("/checkDelivery",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(checkDeliveryHandler)).
		Methods("GET")
		/*
			r.Handle("/payment",
				alice.New(mw.Auth).
					Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
					ThenFunc(paymentHandler)).
				Methods("POST")
		*/
	r.Handle("/placeOrder",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(placeOrderHandler)).
		Methods("POST")

	r.Handle("/status",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(getStatusHandler)).
		Methods("GET")

	r.Handle("/fbtoken",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(fbtokenHandler)).
		Methods("POST")

	r.Handle("/payment-initiate",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole, c.WriterRole, c.UserRole)).
			ThenFunc(initiatePaymentHandler)).
		Methods("POST")

	r.Handle("/forgotPassword",
		alice.New(mw.NoAuth).
			ThenFunc(forgotPasswordHandler)).
		Methods("POST")

	r.Handle("/resetPassword",
		alice.New(mw.NoAuth).
			ThenFunc(resetPasswordHandler)).
		Methods("POST")

	r.Handle("/cache",
		alice.New(mw.Auth).
			ThenFunc(cacheHandler)).
		Methods("GET")

	r.Handle("/cache",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(urlHandler)).
		Methods("POST")

	r.Handle("/posts",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(getPostsHandler)).
		Methods("GET")

	r.Handle("/deletePost",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(deletePostHandler)).
		Methods("POST")

	r.HandleFunc("/vr", vrHandler)

	r.Handle("/initiateSignUp",
		alice.New(mw.NoAuth).
			ThenFunc(initiateSignUpHandler)).
		Methods("POST")

	return mw.Common(r)
}

func initServer() error {
	// Check if tables are created in the database
	err := datastore.InitDb()
	if err != nil {
		return err
	}

	//arr := []int{1}
	//queue.InitQueue(arr)

	return nil
}

func main() {

	aws.DisableModule = false
	aws.DisableSmsModule = false

	initLogging()
	datastore.InitMySql()
	defer datastore.CloseMySql()

	err := initServer()
	if err != nil {
		panic(err)
	}

	originsOk := handlers.AllowedOrigins([]string{"*"})
	headersOk := handlers.AllowedHeaders([]string{"content-type"})
	credsOk := handlers.AllowCredentials()

	log.Info("Server running on port 9980")

	err = http.ListenAndServe(":9980", handlers.CORS(originsOk, headersOk, credsOk)(getRouter()))
	if err != nil {
		log.Error(err)
	}
}
