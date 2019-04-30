package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"rob/lib/datastore"
	mw "rob/lib/middleware"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"rob/lib/queue"
	"strconv"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

var router http.Handler

func TestMain(m *testing.M) {
	beforeTests()
	code := m.Run()
	afterTests()

	os.Exit(code)
}

// Makes sure the test database is created
func TestDatabaseExists(t *testing.T) {
	query := fmt.Sprintf(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name = '%s'`,
		c.TestMysqlDbName)

	if err := notEmptyQuery(query); err != nil {
		t.Error(err)
	}
}

// Makes sure that all the required database tables are created in new db
func TestDatabaseTablesExists(t *testing.T) {
	//reqTables := []string{"Users", "Roles", "UserRole", "Mascot", "PostQueue", "Sale"}

	for _, reqTable := range c.MysqlTables {
		query := fmt.Sprintf(`
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = '%s'
			AND table_name = '%s';`, c.TestMysqlDbName, reqTable)

		if err := notEmptyQuery(query); err != nil {
			t.Error(fmt.Sprintf("Table %q not found", reqTable), err)
		}
	}
}

func notEmptyQuery(query string) error {
	db, err := sql.Open("mysql", c.DbUriBase)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Atleast one row should be present
	if rows.Next() == false {
		return errors.New("Empty result set")
	}

	return nil
}

type RolePair struct {
	Name string
	Id   int
}

var rolePairs = []RolePair{
	{c.AdminRoleName, c.AdminRole},
	{c.UserRoleName, c.UserRole},
	{c.WriterRoleName, c.WriterRole},
}

type AccessTest struct {
	Endpoint       string
	Method         string
	Authentication bool
	Authorization  []int
}

// Authentication test for all end points
func TestAccess(t *testing.T) {
	allRoles := []int{c.AdminRole, c.UserRole, c.WriterRole}

	endpoints := []AccessTest{
		{
			"/login",
			http.MethodPost,
			false,
			nil,
		},
		{
			"/signup",
			http.MethodPost,
			false,
			nil,
		},
		{
			"/post",
			http.MethodPost,
			true,
			[]int{c.AdminRole, c.WriterRole},
		},
		{
			"/postlink",
			http.MethodPost,
			true,
			[]int{c.AdminRole, c.WriterRole},
		},
		{
			"/editprofile",
			http.MethodPost,
			true,
			allRoles,
		},
		{
			"/feed",
			http.MethodPost,
			true,
			allRoles,
		},
		{
			"/product",
			http.MethodPost,
			true,
			[]int{c.AdminRole, c.WriterRole},
		},
		{
			"/product",
			http.MethodGet,
			true,
			allRoles,
		},
		{
			"/sale",
			http.MethodPost,
			true,
			[]int{c.AdminRole, c.WriterRole},
		},
		{
			"/sale",
			http.MethodGet,
			true,
			allRoles,
		},
		{
			"/sales",
			http.MethodGet,
			true,
			allRoles,
		},
		{
			"/feedback",
			http.MethodPost,
			true,
			allRoles,
		},
		{
			"/logout",
			http.MethodGet,
			true,
			allRoles,
		},
		// Always the last endpoint should be logout
	}

	// Test all no auth requests
	for _, e := range endpoints {
		req, _ := http.NewRequest(e.Method, e.Endpoint, nil)
		res := executeRequest(req)

		if e.Authentication == false && res.Code == http.StatusUnauthorized {
			// No auth, so the code should not be 403
			t.Errorf("For endpoint=%s, method=%s, no auth, expected!=401 but received=%d", e.Endpoint, e.Method, res.Code)
		}
		if e.Authentication == true && res.Code != http.StatusUnauthorized {
			// Auth, so the code should be 403
			t.Errorf("For endpoint=%s, method=%s, req auth, expected=401 but received=%d", e.Endpoint, e.Method, res.Code)

		}
	}

	// Login as each user and test access
	for _, role := range rolePairs {
		// login as this user
		ph := testPhone(role.Name)
		pw := testPassword(role.Name)
		loginCookie, err := loginUser(ph, pw)
		if err != nil {
			t.Fatal(err)
		}
		// login success, now try the endpoints
		for _, e := range endpoints {
			if e.Authentication {
				ireq, _ := http.NewRequest(e.Method, e.Endpoint, nil)
				ireq.Header.Set("Cookie", loginCookie)
				ires := executeRequest(ireq)

				shouldBeValid := inArray(role.Id, e.Authorization)

				if shouldBeValid && ires.Code == http.StatusUnauthorized {
					// Should have been authorized
					t.Errorf("For endpoint=%s, method=%s, user=%s, expected != 401 but received=%d", e.Endpoint, e.Method, ph, ires.Code)
				}

				if !shouldBeValid && ires.Code != http.StatusUnauthorized {
					// should have been blocked
					t.Errorf("For endpoint=%s, method=%s, user=%s, expected = 401 but received=%d", e.Endpoint, e.Method, ph, ires.Code)
				}

			}
		}

		// logout the user
		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		_ = executeRequest(req)
	}
}

// Tests the sign up endpoint
// First, tries to login as 'signup@example' user and should get error
// Then, signup is done for that user
// Then, login is tried with wrong password
// Then, correct login is done
func TestSignUp(t *testing.T) {
	sn := "signup"
	ph := testPhone(sn)
	name := testName(sn)
	pw := testPassword(sn)
	c.EV = false
	_, err := loginUser(ph, pw)

	if err == nil {
		t.Fatalf("Expected user %q login to fail but it is success", ph)
	}

	err = initiateSignUp(ph)

	if err != nil {
		t.Fatalf("Failed to initiateSignUp for %s", ph)
	}

	u, err := datastore.GetUserByPhone(ph)
	if err != nil {
		t.Fatalf("Failed to get userByPhone for %s", ph)
	}

	err = signUpUser(ph, u.Code, name, pw)
	if err != nil {
		t.Fatal("Signup of user failed", err)
	}

	wpw := strings.ToLower(pw)
	_, err = loginUser(ph, wpw)
	if err == nil {
		t.Errorf("Expected user %q login to fail with password %q but it is success", ph, wpw)
	}

	_, err = loginUser(ph, pw)
	if err != nil {
		t.Errorf("Expected user %q login to be success with password %q but it failed with error=%s", ph, pw, err.Error())
	}
}

// Tests the product endpoints
// All the tables are cleared to make sure they are in empty state
// Tries to fetch product details of an invalid product id and should get error
// Then we create 10 products. Assert success
// Then we request to fetch details of the products added above. Assert sucess
func TestProduct(t *testing.T) {
	// START STEP 1
	// clear mongo product table
	session, err := mgo.Dial(c.Server)
	if err != nil {
		t.Fatal("Failed to connect to mongo", err)
	}
	defer session.Close()

	co := session.DB(c.DbName).C(c.ProductCollection)

	_, err = co.RemoveAll(nil)
	if err != nil {
		t.Fatal("Failed clearing mongo", err)
	}

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// Check with an non existing product id
	var randomId = "59969fce895a1d431178cc9c"
	var endpoint = "/product" + "?pId=" + randomId
	code := getRequest(endpoint, t, loginCookie)
	if code != http.StatusNotFound {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=404 but received=%d", endpoint, "GET", code)
	}

	// Check with an invalid product id
	var invalidId = "5999c"
	endpoint = "/product" + "?pId=" + invalidId

	code = getRequest(endpoint, t, loginCookie)

	if code != http.StatusInternalServerError {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=500 but received=%d", endpoint, "GET", code)
	}

	// Create products and assert success

	totalProducts := 10

	allProducts := make([]types.Product, totalProducts)

	for i := 0; i < totalProducts; i++ {
		allProducts[i].Sku = fmt.Sprintf("Sku_%d", i)
		allProducts[i].Title = fmt.Sprintf("Title_%d", i)
		allProducts[i].Brand = fmt.Sprintf("Brand_%d", i)
		allProducts[i].Quantity = 10
		allProducts[i].UnitPrice = 10
		allProducts[i].Summary = fmt.Sprintf("Summary_%d", i)
		allProducts[i].Image = fmt.Sprintf("http://img-%d.com", i)
		allProducts[i].ThumbNail = fmt.Sprintf("http://thumb_img-%d.com", i)
		allProducts[i].Description = fmt.Sprintf("Description_%d", i)
		allProducts[i].Color = fmt.Sprintf("Color_%d", i)
		allProducts[i].Size = fmt.Sprintf("Size_%d", i)
		id := createProduct(allProducts[i], t, loginCookie)
		allProducts[i].Id = bson.ObjectIdHex(id)

	}

	// Assert all prpducts are in the database
	for i := 0; i < totalProducts; i++ {
		_, err := datastore.GetProduct(allProducts[i].Id.Hex())
		if err != nil {
			t.Fatal("Failed to retrive created product ", i, err)
			return
		}

	}

}

// Tests Sales related endpoints
//1. Makes sure the Sale table is empty
//2. Tries to get sale info with a non existing saleid
//3. Creates 10 sales and asserts
//4. Gets info for all the sales created above and asserts
func TestSale(t *testing.T) {
	//1. Start step 1
	// clear mysql Sale table
	db, err := sql.Open("mysql", c.DbUri)
	if err != nil {
		t.Fatal("Failed to connect to mysql", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`
		TRUNCATE %s`,
		c.SaleTable)

	err = datastore.PrepareAndExec(query, db)
	if err != nil {
		t.Fatal("Failed to clear mysql table", err)
	}

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// Check with an non existing sale id
	var randomId = "123213"
	var endpoint = "/sale" + "?sId=" + randomId
	code := getRequest(endpoint, t, loginCookie)
	if code != http.StatusNotFound {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=404 but received=%d", endpoint, "GET", code)
	}

	// Check with an invalid sale id
	var invalidId = "5999c"
	endpoint = "/sale" + "?sId=" + invalidId

	code = getRequest(endpoint, t, loginCookie)

	if code != http.StatusBadRequest {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=400 but received=%d", endpoint, "GET", code)
	}

	// START STEP 3
	// Create  sale and assert success

	totalSales := 10

	allSales := make([]types.Sale, totalSales)

	for i := 0; i < totalSales; i++ {
		allSales[i].Title = fmt.Sprintf("Title_%d", i)
		allSales[i].Brand = fmt.Sprintf("Brand_%d", i)
		allSales[i].ProductSku = fmt.Sprintf("ProductSku_%d", i)
		allSales[i].Description = fmt.Sprintf("Description_%d", i)
		allSales[i].ThumbNail = fmt.Sprintf("ThumbNail_%d", i)
		allSales[i].StockUnits = 100
		allSales[i].SaleStartTime = 1503043001979004500
		allSales[i].SaleEndTime = 1503043031783511500

		id := createSale(allSales[i], t, loginCookie)
		allSales[i].Id, err = strconv.Atoi(id)
	}

	// Assert all sale are in the database
	for i := 0; i < totalSales; i++ {
		s, err := getSale(allSales[i].Id, t, loginCookie)
		if err != nil {
			t.Fatal("Failed to retrive created sale ", i, err)
			return
		}
		if !compareSale(*s, allSales[i], t) {
			t.Errorf("Sale data and response from endpoint fetch Mismatch. \n For requested sale id = %d , found %d", allSales[i].Id, s.Id)
		}

	}

	// Assert /sales has a non empty output
	sales, err := datastore.GetSales()
	if err != nil {
		t.Fatal("Failed to retrieve the Sales info. ", err)
		return
	}
	if len(sales.Data) == 0 {
		t.Errorf("Sales List is Empty")
	}

}

/*
type addPostLink struct {
	pIndex         int
	mascotId       int
	mascot11PIndex []int
	mascot12PIndex []int
}
*/
// Tests for placing an order
// 1. All the tables and cleared to make ensure they are empty
// 2. Place an order for a non existing product . Assert failure
// 3. Place an order for a out of stock product . Assert failure
// 4. Place 10 orders with 10 addresses . assert success
// 5. Retrieve all orders placed by the user . Assert success

func TestOrder(t *testing.T) {
	//1. Start step 1
	// clear mysql Orders table
	db, err := sql.Open("mysql", c.DbUri)
	if err != nil {
		t.Fatal("Failed to connect to mysql", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`
		TRUNCATE %s`,
		c.OrderTable)

	err = datastore.PrepareAndExec(query, db)
	if err != nil {
		t.Fatal("Failed to clear mysql table", err)
	}

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// Create a order with an invalid productId
	var randomProductId = "123213"
	var endpoint = "/order"
	order := orderInIt(randomProductId, 1503042122863263469, 100, 100, 0, 1, 200, 1)
	code := createOrderWithStatus(order, t, loginCookie)
	if code != http.StatusBadRequest {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=400 but received=%d", endpoint, "POST", code)
	}

	// create a order with non existing product id
	randomProductId = "59969fce895a1d431178cc9c"
	endpoint = "/order"
	order = orderInIt(randomProductId, 1503042122863263469, 100, 100, 0, 1, 200, 1)
	code = createOrderWithStatus(order, t, loginCookie)
	if code != http.StatusNotFound {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=404 but received=%d", endpoint, "POST", code)
	}

	var product types.Product
	product.Sku = fmt.Sprintf("Sku")
	product.Title = fmt.Sprintf("Title")
	product.Brand = fmt.Sprintf("Brand")
	product.Quantity = 0
	product.UnitPrice = 10
	product.Summary = fmt.Sprintf("Summary")
	product.Image = fmt.Sprintf("http://img.com")
	product.ThumbNail = fmt.Sprintf("http://thumb_img.com")
	product.Description = fmt.Sprintf("Description")
	product.Color = fmt.Sprintf("Color")
	product.Size = fmt.Sprintf("Size")
	outOfStockId := createProduct(product, t, loginCookie)

	// create a order with out of stock product id
	endpoint = "/order"
	order = orderInIt(outOfStockId, 1503042122863263469, 100, 100, 0, 1, 200, 1)
	code = createOrderWithStatus(order, t, loginCookie)
	if code != http.StatusNotFound {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=404 but received=%d", endpoint, "POST", code)
	}

	// Create addresses and assert success
	var ids []int
	totalAddresses := 10
	var address types.Address
	for j := 0; j < totalAddresses; j++ {
		address.Address = fmt.Sprintf("Address_%d", j)
		address.AddressType = fmt.Sprintf("Address Type_%d", j)
		address.City = fmt.Sprintf("City_%d", j)
		address.State = fmt.Sprintf("State_%d", j)
		address.PostalCode = 560066
		address.Phone = fmt.Sprintf("809588337%d", j)
		id, _ := strconv.Atoi(createAddress(address, t, loginCookie))
		ids = append(ids, id)
	}

	// START STEP 3
	// Create  orders and assert success
	totalOrders := 10
	product.Quantity = 100
	productId := createProduct(product, t, loginCookie)
	var allOrders = make([]types.Order, totalOrders)
	for i := 0; i < totalOrders; i++ {

		allOrders[i] = orderInIt(productId, 1503043001979004500, (100 + i), (50 + i), 0, ids[i], 200, 1)
		allOrders[i].ProductTitle = product.Title
		allOrders[i].ProductThumb = product.ThumbNail
		allOrders[i].TransId = c.UninitiatedId
		allOrders[i].TransStatus = c.Uninitiated
		allOrders[i].ShippingStatus = c.Uninitiated
		allOrders[i].TrackingId = c.Uninitiated
		allOrders[i].ShippingId = c.UninitiatedId
		allOrders[i].Id, _ = strconv.Atoi(createOrder(allOrders[i], t, loginCookie))

	}
	// Assert all orders are in the database
	for i := 0; i < totalOrders; i++ {
		o, err := getOrder(allOrders[i].Id, t, loginCookie)
		if err != nil {
			t.Fatal("Failed to retrive created order ", i, err)
			return
		}
		o.TimeOfCreation = int64(0)
		allOrders[i].TimeOfCreation = int64(0)
		if !compareOrder(*o, allOrders[i], t) {
			t.Errorf("Order data and response from endpoint fetch Mismatch. \n For requested order id = %d , found %d", allOrders[i].Id, o.Id)
		}

	}

	// Assert /orders has a 10 output
	orders, err := getOrders(t, loginCookie)
	if err != nil {
		t.Fatal("Failed to retrieve the orders info. ", err)
		return
	}
	if len(orders.Data) != 10 {
		t.Errorf("Number of orders added do not matches number of orders retrieved")
	}

}

// Tests related to address
//1. All the tables are cleared to make sure they are empty
//2. Create address with invalid postal code , assert failure
//3. Create 10 addresses . Assert success
//4. Get all addresses created above . Assert success

func TestAddress(t *testing.T) {

	//1. Start step 1
	// clear mysql Address table
	db, err := sql.Open("mysql", c.DbUri)
	if err != nil {
		t.Fatal("Failed to connect to mysql", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`
		TRUNCATE %s`,
		c.AddressTable)

	err = datastore.PrepareAndExec(query, db)
	if err != nil {
		t.Fatal("Failed to clear mysql table", err)
	}

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// Check with an non existing address id
	var randomId = "123213"
	var endpoint = "/address" + "?aId=" + randomId
	code := getRequest(endpoint, t, loginCookie)
	if code != http.StatusNotFound {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=404 but received=%d", endpoint, "GET", code)
	}

	// Check with an invalid address id
	var invalidId = "5999c"
	endpoint = "/address" + "?aId=" + invalidId

	code = getRequest(endpoint, t, loginCookie)

	if code != http.StatusBadRequest {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=400 but received=%d", endpoint, "GET", code)
	}

	// START STEP 3
	// Create  sale and assert success

	totalAddresses := 10

	allAddresses := make([]types.Address, totalAddresses)

	for i := 0; i < totalAddresses; i++ {
		allAddresses[i].Address = fmt.Sprintf("Address_%d", i)
		allAddresses[i].AddressType = fmt.Sprintf("Address Type_%d", i)
		allAddresses[i].City = fmt.Sprintf("City_%d", i)
		allAddresses[i].State = fmt.Sprintf("State_%d", i)
		allAddresses[i].PostalCode = 560066
		allAddresses[i].Phone = fmt.Sprintf("809588337%d", i)
		id := createAddress(allAddresses[i], t, loginCookie)
		allAddresses[i].Id, _ = strconv.Atoi(id)
	}

	// Assert all address are in the database
	for i := 0; i < totalAddresses; i++ {
		a, err := getAddress(allAddresses[i].Id, t, loginCookie)
		if err != nil {
			t.Fatal("Failed to retrive created address ", i, err)
			return
		}
		a.TimeOfCreation = int64(0)
		allAddresses[i].TimeOfCreation = int64(0)
		if !compareAddress(*a, allAddresses[i], t) {
			t.Errorf("Address data and response from endpoint fetch Mismatch. \n For requested address id = %d , found %d", allAddresses[i].Id, a.Id)
		}

	}

	// Edit the address and assert
	allAddresses[0].Address = "New Address"
	err = editAddress(allAddresses[0].Id, allAddresses[0], loginCookie)
	if err != nil {
		log.Fatal("Couldn't Edit address")
	}
	new, err := getAddress(allAddresses[0].Id, t, loginCookie)
	if err != nil {
		t.Fatal("Failed to retrive edited address ", err)
		return
	}
	if new.Address != "New Address" {
		t.Error("Edit data mismatch")
	}

	// Assert /addresses has a non empty output
	addresses, err := getAddresses(t, loginCookie)
	if err != nil {
		t.Fatal("Failed to retrieve the Addresses info. ", err)
		return
	}
	if len(addresses.Data) == 0 {
		t.Errorf("Addresses List is Empty")
	}
}

// Test for Delivary location check
// 1. Check for a invalid postal code . assert failure
// 2. Check for a non deliverable postal code . Assert failure
// 3. Check for 10 deliverable postal code . Assert success

func TestDelivery(t *testing.T) {

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// 1. Invalid postal code
	invalidId := "5600yy"
	endpoint := "/checkDelivery?pc=" + invalidId
	code := getRequest(endpoint, t, loginCookie)
	if code != http.StatusBadRequest {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=400 but received=%d", endpoint, "GET", code)
	}

	// 2. Non Deliverable postal code
	ndId := "450066"
	endpoint = "/checkDelivery?pc=" + ndId
	message, err := checkDelivery(endpoint, t, loginCookie)
	if message != "Not Deliverable" {
		t.Errorf("For endpoint=%s, method=%s, auth, expected=Not Deliverable but received=%s", endpoint, "GET", message)
	}

	// 3.Check 10 valid delivery
	totalCodes := 10
	for i := 0; i < totalCodes; i++ {
		endpoint = "/checkDelivery?pc=" + "56006" + strconv.Itoa(i)
		code = getRequest(endpoint, t, loginCookie)
		if code != http.StatusOK {
			t.Errorf("For endpoint=%s, method=%s, auth, expected=200 but received=%d", endpoint, "GET", code)
		}
	}

}

func pickPosts(all []types.Post, sel []int) []types.Post {
	var r []types.Post
	for _, i := range sel {
		r = append(r, all[i])
	}
	return r
}

// Post, postlink, feed tests
// 1. All the tables are cleared to make sure they are in empty state
// 2. Create mascots 11, 12
// 3. Assert their feed is empty
// 4. Create 10 posts. Assert success
// 5. Assert mascot feeds are empty again
// 6. Create postlink with non existing post id, assert failure
// 7. Create postlink with non existing mascot id, assert failure
// 8. Create 12 postlink, assert feed after every addition
// 9. assert feed before,after and with 0
// 10. Do feed testing with multiple timestamps
func TestFeed(t *testing.T) {

	// START STEP 1
	// clear mongo post table

	session, err := mgo.Dial(c.Server)
	if err != nil {
		t.Fatal("Failed to connect to mongo", err)
	}
	defer session.Close()

	co := session.DB(c.DbName).C(c.Collection)

	_, err = co.RemoveAll(nil)
	if err != nil {
		t.Fatal("Failed clearing mongo", err)
	}

	// clear mysql postqueue table
	db, err := sql.Open("mysql", c.DbUri)
	if err != nil {
		t.Fatal("Failed to connect to mysql", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`
		TRUNCATE %s`,
		c.PostQueueTable)

	err = datastore.PrepareAndExec(query, db)
	if err != nil {
		t.Fatal("Failed to clear mysql table", err)
	}

	// START STEP 2
	// create mascots 11, 12

	mascot11Name := "mascot11"
	mascot12Name := "mascot12"

	mascot11Id := 11
	mascot12Id := 12

	query = fmt.Sprintf(`
		INSERT INTO %s
		VALUES(%d,'%s','%s')
		ON DUPLICATE KEY
		UPDATE %s = %d;`,
		c.MascotTable, mascot11Id, mascot11Name, "",
		c.Id, mascot11Id)

	if err := datastore.PrepareAndExec(query, db); err != nil {
		t.Fatal("Failed to create mascot 11", err)
	}

	query = fmt.Sprintf(`
		INSERT INTO %s
		VALUES(%d,'%s','%s')
		ON DUPLICATE KEY
		UPDATE %s = %d;`,
		c.MascotTable, mascot12Id, mascot12Name, "",
		c.Id, mascot12Id)

	if err := datastore.PrepareAndExec(query, db); err != nil {
		t.Fatal("Failed to create mascot 12", err)
	}

	// START STEP 3

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// Assert mascots have empty feeds

	emptyFeed := []types.Post{}

	compareFeed(mascot11Id, -1, 1, emptyFeed, t, loginCookie)
	compareFeed(mascot12Id, -1, 1, emptyFeed, t, loginCookie)

	compareFeed(mascot11Id, -1, 3, emptyFeed, t, loginCookie)
	compareFeed(mascot12Id, -1, 3, emptyFeed, t, loginCookie)

	// START STEP 4
	// Create a post and assert success

	totalPosts := 10

	allPosts := make([]types.Post, totalPosts)

	for i := 0; i < totalPosts; i++ {

		allPosts[i].CardType = c.CardTypeImage
		allPosts[i].Title = fmt.Sprintf("Title_%d", i)
		allPosts[i].DpSrc = "test"
		allPosts[i].Src = fmt.Sprintf("http://img-%d.com", i)
		allPosts[i].Description = fmt.Sprintf("Description_%d", i)
		allPosts[i].ButtonText = fmt.Sprintf("ButtonText_%d", i)
		allPosts[i].Url = fmt.Sprintf("Url_%d", i)
		allPosts[i].CardType = 1
		allPosts[i].ChildPosts = []string{}
		if i == 2 || i == 3 {
			allPosts[i].CardType = c.CardTypeList
			allPosts[i].ChildPosts = []string{allPosts[0].Id.Hex(), allPosts[1].Id.Hex()}
			allPosts[i].GradientStart = "#123456"
			allPosts[i].GradientEnd = "#123456"
			allPosts[i].Icon = fmt.Sprintf("Icon_%d", i)
		}
		id := createPost(allPosts[i], t, loginCookie)
		allPosts[i].Id = bson.ObjectIdHex(id)
		// add new postlink
		if i%2 == 0 {
			err := createPostLink(allPosts[i].Id.Hex(), mascot11Id, loginCookie)
			if err != nil {
				t.Fatal("Failed to create postlink. index=", i, err)
			}
		} else {
			// add new postlink
			err := createPostLink(allPosts[i].Id.Hex(), mascot12Id, loginCookie)
			if err != nil {
				t.Fatal("Failed to create postlink. index=", i, err)
			}
		}

	}

	// Assert all posts are in the database
	for i := 0; i < totalPosts; i++ {
		p, err := datastore.GetPostMetaData(allPosts[i].Id.Hex())
		if err != nil {
			t.Fatal("Failed to retrive created post ", i, err)
		}
		allPosts[i].TimeOfCreation = p.TimeOfCreation
	}

	// START STEP 6
	// Create postlink with non existing post id and assert failure
	invalidPostId := bson.NewObjectId().Hex()
	invalidMascotId := 1024
	err = createPostLink(invalidPostId, mascot11Id, loginCookie)
	if err == nil {
		t.Fatalf("Expected creation of postlink to fail with postId=%s, mascotId=%d but it passed", invalidPostId, mascot11Id)
	}

	// START STEP 7
	// Create postlink with non existing mascot id and assert failure

	err = createPostLink(allPosts[0].Id.Hex(), invalidMascotId, loginCookie)
	if err == nil {
		t.Fatalf("Expected creation of postlink to fail with postId=%s, mascotId=%d but it passed", allPosts[0].Id.Hex(), invalidMascotId)
	}

	postlinks11, _ := datastore.GetPostLinks(mascot11Id)
	postlinks12, _ := datastore.GetPostLinks(mascot12Id)

	//set ChildPostsJson
	for i := 0; i < totalPosts; i++ {
		allPosts[i].ChildPostsJson = ""
		if i == 2 || i == 3 {
			if i%2 == 0 {
				allPosts[0].TimeOfLink = postlinks11[i/2].TimeOfCreation
				allPosts[1].TimeOfLink = postlinks11[i/2].TimeOfCreation
			} else {
				allPosts[0].TimeOfLink = postlinks12[i/2].TimeOfCreation
				allPosts[1].TimeOfLink = postlinks12[i/2].TimeOfCreation
			}
			m, err := json.Marshal([]types.Post{allPosts[0], allPosts[1]})
			if err != nil {
				t.Fatal(err)
			}
			allPosts[i].ChildPostsJson = string(m)
		}
	}

	for i := 0; i < totalPosts; i++ {
		if i%2 == 0 {
			allPosts[i].TimeOfLink = postlinks11[i/2].TimeOfCreation
		} else {
			allPosts[i].TimeOfLink = postlinks12[i/2].TimeOfCreation
		}
	}

	//TopPosts
	compareFeed(mascot11Id, -1, 1, pickPosts(allPosts, []int{8, 6, 4, 2, 0}), t, loginCookie)
	compareFeed(mascot12Id, -1, 1, pickPosts(allPosts, []int{9, 7, 5, 3, 1}), t, loginCookie)

	timestamp11 := postlinks11[2].TimeOfCreation
	timestamp12 := postlinks12[2].TimeOfCreation

	fmt.Println("STARTING POSTSBEFORE")

	//PostsBefore timestamp
	compareFeed(mascot11Id, timestamp11, 2, pickPosts(allPosts, []int{2, 0}), t, loginCookie)
	compareFeed(mascot12Id, timestamp12, 2, pickPosts(allPosts, []int{3, 1}), t, loginCookie)

	fmt.Println("STARTING POSTSAFTER")
	//PostsAfter timestamp
	compareFeed(mascot11Id, timestamp11, 3, pickPosts(allPosts, []int{8, 6}), t, loginCookie)
	compareFeed(mascot12Id, timestamp12, 3, pickPosts(allPosts, []int{9, 7}), t, loginCookie)

	// test for Different NumberOfposts
	c.NumOfPosts = 1
	//TopPosts
	compareFeed(mascot11Id, -1, 1, pickPosts(allPosts, []int{8}), t, loginCookie)
	compareFeed(mascot12Id, -1, 1, pickPosts(allPosts, []int{9}), t, loginCookie)

	//PostsBefore timestamp
	compareFeed(mascot11Id, timestamp11, 2, pickPosts(allPosts, []int{2}), t, loginCookie)
	compareFeed(mascot12Id, timestamp12, 2, pickPosts(allPosts, []int{3}), t, loginCookie)

	//PostsAfter timestamp
	compareFeed(mascot11Id, timestamp11, 3, pickPosts(allPosts, []int{6}), t, loginCookie)
	compareFeed(mascot12Id, timestamp12, 3, pickPosts(allPosts, []int{7}), t, loginCookie)

}

// 1. Create feedback using endpoint
// 2. Test that database has the contents
func TestFeedback(t *testing.T) {
	ph := testPhone(c.UserRoleName)
	pw := testPassword(c.UserRoleName)
	loginCookie, err := loginUser(ph, pw)
	if err != nil {
		t.Fatal("User login failed", err)
	}

	err = createFeedback(c.Feedback, "test desc", loginCookie)
	if err != nil {
		t.Fatalf("Expected creation of new feedback to pass but it failed")
	}

	// Assert the item is in database
	query := fmt.Sprintf(`
		SELECT *
		FROM %s.%s
		WHERE %s = '%s'`,
		c.TestMysqlDbName, c.FeedbackTable,
		c.Phone, ph)

	err = notEmptyQuery(query)
	if err != nil {
		t.Error(err)
	}
}

// Tests related to payment
// 1. Clear all tables
//2 . Try to process payment for an invalid orderid . Assert failure
// 3. Try to process payment for an non existing orderid. Assert failure
// 7. Try to process payment for an out of stock product for a orderid . Assert failure
// 8. Process a valid payment . Assert success

func BrokenTestPayment(t *testing.T) {
	//1. Start step 1
	// clear mysql Transaction table
	db, err := sql.Open("mysql", c.DbUri)
	if err != nil {
		t.Fatal("Failed to connect to mysql", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`
		TRUNCATE %s`,
		c.TransactionTable)

	err = datastore.PrepareAndExec(query, db)
	if err != nil {
		t.Fatal("Failed to clear mysql table", err)
	}

	// Login as admin and use this for all subsequent requests
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}

	// test with invalid orderId
	invalidOrderId := "444a"
	amount := "100"
	code := createPaymentRequest(amount, invalidOrderId, loginCookie, t)
	if code != http.StatusBadRequest {
		t.Errorf("For endpoint=/payment, method=%s, auth, expected=400 but received=%d", "POST", code)
	}

	// test with non existing orderId
	invalidOrderId = "444678"
	amount = "200"
	code = createPaymentRequest(amount, invalidOrderId, loginCookie, t)
	if code != http.StatusNotFound {
		t.Errorf("For endpoint=/payment, method=%s, auth, expected=404 but received=%d", "POST", code)
	}

	// create a sale with 1 stock item
	var sale types.Sale
	sale.Title = fmt.Sprintf("Title")
	sale.Brand = fmt.Sprintf("Brand")
	sale.ProductSku = fmt.Sprintf("ProductSku")
	sale.Description = fmt.Sprintf("Description")
	sale.ThumbNail = fmt.Sprintf("http://localhost")
	sale.StockUnits = 1
	sale.SaleStartTime = 1503043001979004500
	sale.SaleEndTime = 1503043031783511500

	id := createSale(sale, t, loginCookie)
	sale.Id, _ = strconv.Atoi(id)

	// create a product with 1 stock item
	var product types.Product
	product.Sku = fmt.Sprintf("ProductSku")
	product.Title = fmt.Sprintf("Title")
	product.Brand = fmt.Sprintf("Brand")
	product.Quantity = 1
	product.UnitPrice = 10
	product.Summary = fmt.Sprintf("Summary")
	product.Image = fmt.Sprintf("http://img.com")
	product.ThumbNail = fmt.Sprintf("http://thumb_img.com")
	product.Description = fmt.Sprintf("Description")
	product.Color = fmt.Sprintf("Color")
	product.Size = fmt.Sprintf("Size")
	productId := createProduct(product, t, loginCookie)

	// create a valid addresss
	var address types.Address
	address.Address = fmt.Sprintf("Address")
	address.AddressType = fmt.Sprintf("Address Type")
	address.City = fmt.Sprintf("City")
	address.State = fmt.Sprintf("State")
	address.PostalCode = 560066
	address.Phone = fmt.Sprintf("8095883377")
	address.Id, _ = strconv.Atoi(createAddress(address, t, loginCookie))

	// create an order with above data
	var order types.Order
	order = orderInIt(productId, 1503043001979004500, 100, 50, 0, address.Id, 200, sale.Id)
	order.ProductTitle = product.Title
	order.ProductThumb = product.ThumbNail
	order.TransId = c.UninitiatedId
	order.TransStatus = c.Uninitiated
	order.ShippingStatus = c.Uninitiated
	order.TrackingId = c.Uninitiated
	order.ShippingId = c.UninitiatedId
	order.Id, _ = strconv.Atoi(createOrder(order, t, loginCookie))

	// process payment for order
	res, err := createPayment(amount, loginCookie, order.Id, t)
	if err != nil {
		t.Fatal("couldnt process payment", err)
		return
	}
	// Check transaction entry in Transaction table
	exists, err := mw.CheckExistanceMysql(c.TransactionTable, c.Id, strconv.Itoa(res.TransId), true)
	if err != nil {
		t.Fatal("couldnt check transaction entry", err)
		return
	}

	if !exists {
		t.Error("No entry found for the transaction id")
		return
	}
	// Assert that the product and sale have gone of stock
	newProduct, err := datastore.GetProduct(productId)
	if newProduct.Quantity != 0 {
		t.Error("Product Quantity not decremented")
	}
	newSale, err := datastore.GetSale(sale.Id)
	if newSale.StockUnits != 0 {
		t.Error("Sale Stock not decremented")
	}

}

// Test for updating the Instance Id (firebase token ) for a user

func TestToken(t *testing.T) {

	// Login as admin
	loginCookie, err := loginUser(testPhone(c.AdminRoleName), testPassword(c.AdminRoleName))
	if err != nil {
		t.Fatal("Admin login failed", err)
	}
	token := "thisIsATestToken11234567!@#$"
	code := updateToken(token, loginCookie)
	if code != http.StatusOK {
		t.Fatal("Updating token failed")
	}
}

func updateToken(token string, loginCookie string) int {
	data := url.Values{}
	data.Set(c.Token, token)
	req, _ := http.NewRequest(http.MethodPost, "/fbtoken", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	return res.Code
}

func createPaymentRequest(amount, orderId, loginCookie string, t *testing.T) int {
	data := url.Values{}
	data.Set(c.Amount, amount)
	data.Set(c.OrderId, orderId)
	req, _ := http.NewRequest(http.MethodPost, "/payment", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	return res.Code
}

func createPayment(amount, loginCookie string, orderId int, t *testing.T) (*types.TransactionResponse, error) {
	data := url.Values{}
	data.Set(c.Amount, amount)
	id := strconv.Itoa(orderId)
	data.Set(c.OrderId, id)
	req, _ := http.NewRequest(http.MethodPost, "/payment", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)

	if res.Code != http.StatusOK {
		t.Fatalf("processing payment request failed for orderId=%d", orderId)
		return nil, errors.New("processing payment request returned non-OK")
	}
	var result types.TransactionResponse
	log.Debug("response from payment processing %s", res.Body)
	err := json.Unmarshal(res.Body.Bytes(), &result)
	if err != nil {
		t.Fatal("Couldnt marshal payment response", err)
		return nil, err
	}
	return &result, nil
}

func createFeedback(typ, desc, loginCookie string) error {
	data := url.Values{}
	data.Set(c.Type, typ)
	data.Set(c.Description, desc)
	req, _ := http.NewRequest(http.MethodPost, "/feedback", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		return errors.New(fmt.Sprintf("Status not Ok on feedback creation"))
	}
	return nil
}

func createPostLink(postId string, mascotId int, loginCookie string) error {
	data := url.Values{}
	data.Set(c.PostId, postId)
	data.Set(c.MascotId, fmt.Sprintf("%d", mascotId))
	req, _ := http.NewRequest(http.MethodPost, "/postlink", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		return errors.New(fmt.Sprintf("Status not Ok on postlink creation with postId=%s, mascotId=%d", postId, mascotId))
	}
	return nil
}

func createPost(p types.Post, t *testing.T, loginCookie string) string {
	data := url.Values{}
	data.Set(c.CardType, fmt.Sprintf("%d", p.CardType))
	data.Set(c.DpSrc, p.DpSrc)
	data.Set(c.Title, p.Title)
	data.Set(c.Src, p.Src)
	data.Set(c.Description, p.Description)
	data.Set(c.ButtonText, p.ButtonText)
	data.Set(c.Url, p.Url)
	for _, v := range p.ChildPosts {
		data.Add(c.ChildPosts, v)
	}
	data.Set(c.Icon, p.Icon)
	data.Set(c.GradientStart, p.GradientStart)
	data.Set(c.GradientEnd, p.GradientEnd)

	req, _ := http.NewRequest(http.MethodPost, "/post", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Create Post request failed. Title=%s. Status=%d", p.Title, res.Code)
	}
	var id string
	err := json.Unmarshal(res.Body.Bytes(), &id)
	if err != nil {
		t.Fatal("Create Post response unmarshal fail", err)
	}
	return id
}

func createAddress(a types.Address, t *testing.T, loginCookie string) string {
	data := url.Values{}
	data.Set(c.Address, a.Address)
	data.Set(c.AddressType, a.AddressType)
	data.Set(c.City, a.City)
	data.Set(c.State, a.State)
	pc := strconv.Itoa(a.PostalCode)
	data.Set(c.PostalCode, pc)
	data.Set(c.Phone, a.Phone)
	log.Debug("Address data = %+v\n", a)
	req, _ := http.NewRequest(http.MethodPost, "/address", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Create Address request failed.%+v\n", a)
	}
	var id string
	log.Debug("response from add address  %s", res.Body)
	err := json.Unmarshal(res.Body.Bytes(), &id)
	if err != nil {
		t.Fatal("Create address response unmarshal fail", err)
	}
	return id
}

func createProduct(p types.Product, t *testing.T, loginCookie string) string {
	data := url.Values{}
	data.Set(c.Sku, p.Sku)
	data.Set(c.Title, p.Title)
	data.Set(c.Brand, p.Brand)
	quantity := strconv.Itoa(p.Quantity)
	data.Set(c.Quantity, quantity)
	data.Set(c.Description, p.Description)
	unitPrice := strconv.Itoa(p.UnitPrice)
	data.Set(c.UnitPrice, unitPrice)
	data.Set(c.Summary, p.Summary)
	data.Set(c.Image, p.Image)
	data.Set(c.ThumbNail, p.ThumbNail)
	data.Set(c.Color, p.Color)
	data.Set(c.Size, p.Size)
	log.Debug("Product data = %+v\n", p)
	req, _ := http.NewRequest(http.MethodPost, "/product", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Create Product request failed.%+v\n", p)
	}
	var id string
	err := json.Unmarshal(res.Body.Bytes(), &id)
	if err != nil {
		t.Fatal("Create Product response unmarshal fail", err)
	}
	return id
}

func createSale(s types.Sale, t *testing.T, loginCookie string) string {
	data := url.Values{}
	data.Set(c.Title, s.Title)
	data.Set(c.Brand, s.Brand)
	data.Set(c.ProductSku, s.ProductSku)
	data.Set(c.Description, s.Description)
	data.Set(c.ThumbNail, s.ThumbNail)
	stock := strconv.Itoa(s.StockUnits)
	data.Set(c.StockUnits, stock)
	start := strconv.FormatInt(s.SaleStartTime, 10)
	data.Set(c.SaleStartTime, start)
	end := strconv.FormatInt(s.SaleEndTime, 10)
	data.Set(c.SaleEndTime, end)
	log.Debug("Sale data = %+v\n", s)
	req, _ := http.NewRequest(http.MethodPost, "/sale", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Create Sale request failed.%+v\n", s)
	}
	var id string
	log.Debug("response from create sale %s", res.Body)
	err := json.Unmarshal(res.Body.Bytes(), &id)
	if err != nil {
		t.Fatal("Create Sale response unmarshal fail", err)
	}
	return id

}

func createOrder(o types.Order, t *testing.T, loginCookie string) string {
	data := url.Values{}
	data.Set(c.ProductId, o.ProductId)
	orderdate := strconv.FormatInt(o.OrderDate, 10)
	data.Set(c.OrderDate, orderdate)
	price := strconv.Itoa(o.Price)
	data.Set(c.Price, price)
	tax := strconv.Itoa(o.Tax)
	data.Set(c.Tax, tax)
	cost := strconv.Itoa(o.ShippingCost)
	data.Set(c.ShippingCost, cost)
	amount := strconv.Itoa(o.Amount)
	data.Set(c.Amount, amount)
	saleId := strconv.Itoa(o.SaleId)
	data.Set(c.SaleId, saleId)
	addressid := strconv.Itoa(o.AddressId)
	data.Set(c.AddressId, addressid)
	log.Debug("Order data = %+v\n", o)
	req, _ := http.NewRequest(http.MethodPost, "/order", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)

	if res.Code != http.StatusOK {
		t.Fatalf("Create Order request failed.%+v\n", o)
	}
	var id string
	err := json.Unmarshal(res.Body.Bytes(), &id)
	if err != nil {
		t.Fatal("Create Order response unmarshal fail", err)
	}
	return id
}

func createOrderWithStatus(o types.Order, t *testing.T, loginCookie string) int {
	data := url.Values{}
	data.Set(c.ProductId, o.ProductId)
	orderdate := strconv.FormatInt(o.OrderDate, 10)
	data.Set(c.OrderDate, orderdate)
	price := strconv.Itoa(o.Price)
	data.Set(c.Price, price)
	tax := strconv.Itoa(o.Tax)
	data.Set(c.Tax, tax)
	cost := strconv.Itoa(o.ShippingCost)
	data.Set(c.ShippingCost, cost)
	amount := strconv.Itoa(o.Amount)
	data.Set(c.Amount, amount)
	saleId := strconv.Itoa(o.SaleId)
	data.Set(c.SaleId, saleId)
	addressid := strconv.Itoa(o.AddressId)
	data.Set(c.AddressId, addressid)
	log.Debug("Order data = %+v\n", o)
	req, _ := http.NewRequest(http.MethodPost, "/order", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	return res.Code
}

func getOrders(t *testing.T, loginCookie string) (*types.OrdersList, error) {
	endpoint := "/orders"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("failed to create request for getting sales", err)
		return nil, err
	}
	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Getting Orders request failed for ")
		return nil, errors.New("Get Orders endpoint returned non-OK")
	}
	var orders types.OrdersList
	log.Debug("response from get orders %s", res.Body)
	err = json.Unmarshal(res.Body.Bytes(), &orders)
	if err != nil {
		t.Fatal("Get Orders response unmarshal fail", err)
		return nil, err
	}
	return &orders, nil

}

func getAddresses(t *testing.T, loginCookie string) (*types.AddressList, error) {
	endpoint := "/addresses"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("failed to create request for getting sales", err)
		return nil, err
	}
	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Getting Addresses request failed for ")
		return nil, errors.New("Get Addresses endpoint returned non-OK")
	}
	var addresses types.AddressList
	log.Debug("response from get addresses %s", res.Body)
	err = json.Unmarshal(res.Body.Bytes(), &addresses)
	if err != nil {
		t.Fatal("Get Addresses response unmarshal fail", err)
		return nil, err
	}
	return &addresses, nil

}

func getSale(sId int, t *testing.T, loginCookie string) (*types.Sale, error) {
	saleId := strconv.Itoa(sId)
	endpoint := "/sale?sId=" + saleId
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("failed to create request for getting sales", err)
		return nil, err
	}

	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Getting Sale request failed for sId=%d", sId)
		return nil, errors.New("Get Sale endpoint returned non-OK")
	}
	var sale types.Sale
	log.Debug("response from get sale %s", res.Body)
	err = json.Unmarshal(res.Body.Bytes(), &sale)
	if err != nil {
		t.Fatal("Get Sale response unmarshal fail", err)
		return nil, err
	}
	return &sale, nil

}

func getOrder(oId int, t *testing.T, loginCookie string) (*types.Order, error) {
	orderId := strconv.Itoa(oId)
	endpoint := "/order?oId=" + orderId
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("failed to create request for getting orders", err)
		return nil, err
	}

	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Getting Order request failed for oId=%d", oId)
		return nil, errors.New("Get Order endpoint returned non-OK")
	}
	var order types.Order
	log.Debug("response from get order %s", res.Body)
	err = json.Unmarshal(res.Body.Bytes(), &order)
	if err != nil {
		t.Fatal("Get Order response unmarshal fail", err)
		return nil, err
	}
	return &order, nil

}

func getAddress(aId int, t *testing.T, loginCookie string) (*types.Address, error) {
	addressId := strconv.Itoa(aId)
	endpoint := "/address?aId=" + addressId
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("failed to create request for getting addresss", err)
		return nil, err
	}

	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Getting Address request failed for aId=%d", aId)
		return nil, errors.New("Get Address endpoint returned non-OK")
	}
	var address types.Address
	log.Debug("response from get address %s", res.Body)
	err = json.Unmarshal(res.Body.Bytes(), &address)
	if err != nil {
		t.Fatal("Get Address response unmarshal fail", err)
		return nil, err
	}
	return &address, nil

}

func checkDelivery(endpoint string, t *testing.T, loginCookie string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("failed to create request for delivery check", err)
		return "Failed", err
	}

	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Checking Delivery request failed for endpoint=%s", endpoint)
		return "Failed", errors.New("Delivery Check endpoint returned non-OK")
	}
	var message string
	log.Debug("response from delivery check  %s", res.Body)
	err = json.Unmarshal(res.Body.Bytes(), &message)
	if err != nil {
		t.Fatal("Delivery Check response unmarshal fail", err)
		return "Failed", err
	}
	return message, nil

}

func getRequest(endpoint string, t *testing.T, loginCookie string) int {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Cookie", loginCookie)
	res := executeRequest(req)
	return res.Code

}

func compareSale(sale1 types.Sale, sale2 types.Sale, t *testing.T) bool {
	res1, err := json.Marshal(sale1)
	if err != nil {
		t.Error("Could not marshal reponse from GetSale", err)
		return false
	}

	res2, err := json.Marshal(sale2)
	if err != nil {
		t.Error("Could not marshal sale object inserted testing", err)
		return false
	}
	return string(res1) == string(res2)

}

func compareOrder(order1 types.Order, order2 types.Order, t *testing.T) bool {
	res1, err := json.Marshal(order1)
	if err != nil {
		t.Error("Could not marshal reponse from GetOrder", err)
		return false
	}

	res2, err := json.Marshal(order2)
	if err != nil {
		t.Error("Could not marshal order object inserted in testing", err)
		return false
	}
	return string(res1) == string(res2)

}

func editAddress(addressId int, a types.Address, loginCookie string) error {
	data := url.Values{}
	addressid := strconv.Itoa(addressId)
	data.Set(c.AddressId, addressid)
	data.Set(c.Address, a.Address)
	data.Set(c.AddressType, a.AddressType)
	data.Set(c.City, a.City)
	data.Set(c.State, a.State)
	pc := strconv.Itoa(a.PostalCode)
	data.Set(c.PostalCode, pc)
	data.Set(c.Phone, a.Phone)
	log.Debug("Address data = %+v\n", a)
	req, _ := http.NewRequest(http.MethodPost, "/editAddress", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		log.Fatalf("Create Address request failed. %+v", a)
		err := errors.New("Non OK response from editAddress")
		return err
	}
	return nil
}

func compareAddress(address1 types.Address, address2 types.Address, t *testing.T) bool {
	res1, err := json.Marshal(address1)
	if err != nil {
		t.Error("Could not marshal reponse from GetAddress", err)
		return false
	}

	res2, err := json.Marshal(address2)
	if err != nil {
		t.Error("Could not marshal address object inserted testing", err)
		return false
	}
	return string(res1) == string(res2)

}

func compareFeed(mid int, lastSynced int64, flag int, expected []types.Post,
	t *testing.T, loginCookie string) {
	var funcName = "main_test.go:CompareFeed"
	log.WithFields(log.Fields{
		"mid":         mid,
		"lastSynched": lastSynced,
		"flag":        flag,
		"expected":    expected,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	data := url.Values{}
	data.Set(c.MascotId, fmt.Sprintf("%d", mid))
	data.Set(c.LastSync, fmt.Sprintf("%d", lastSynced))
	data.Set(c.Flag, fmt.Sprintf("%d", flag))
	req, _ := http.NewRequest(http.MethodPost, "/feed", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", loginCookie)

	res := executeRequest(req)
	if res.Code != http.StatusOK {
		t.Fatalf("Feed request failed. Mascot=%d. LastSynced=%d. Flag=%d. Status=%d", mid, lastSynced, flag, res.Code)
	}

	gotJson := res.Body.String()

	log.Debugf("Response %v", gotJson)
	expectedJsonB, err := json.Marshal(expected)
	expectedJson := string(expectedJsonB)

	if err != nil {
		t.Fatal("Failed to marshal expected json for mascot ", mid)
	}

	if expectedJson != gotJson {
		t.Fatalf("compare Feed failed for mascot %d\nExpected: %s\nReceived: %s",
			mid, expectedJson, gotJson)
	}

}

func loginUser(ph, pw string) (string, error) {
	data := url.Values{}
	data.Set(c.Phone, ph)
	data.Set(c.Password, pw)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Login of user %q failed", ph))
	}

	loginCookie := res.Header().Get("Set-Cookie")
	index := strings.Index(loginCookie, ";")
	if index == -1 {
		return "", errors.New("Login cookie failed. Does not have ';'")
	}
	loginCookie = loginCookie[0:index]
	return loginCookie, nil
}

func inArray(a int, aa []int) bool {
	for _, b := range aa {
		if a == b {
			return true
		}
	}
	return false
}
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func beforeTests() {
	// Set config to test database
	c.SetMysqlCreds(c.TestMysqlUser, c.TestMysqlPass, c.TestMysqlDbName)
	initLogging()
	// drop database incase previous run panniced and left the db intact
	dropDatabase()
	datastore.InitMySql()
	initServer()
	router = getRouter()
	// Add default users
	for _, role := range rolePairs {
		if err := createUser(role.Name, role.Id); err != nil {
			panic(err)
		}
	}
}

func createUser(s string, r int) error {
	ph := testPhone(s)
	err := initiateSignUp(ph)
	if err != nil {
		return err
	}
	na := testName(s)
	pw := testPassword(s)

	user, err := datastore.GetUserByPhone(ph)
	if err != nil {
		return err
	}

	err = signUpUser(ph, user.Code, na, pw)
	if err != nil {
		return err
	}

	log.Debugf("Created user: %d %s", user.Id, user.Phone)
	err = datastore.UpdateRole(user.Id, r)
	return err
}

func signUpUser(ph, code, name, pw string) error {
	data := url.Values{}
	data.Set(c.Phone, ph)
	data.Set(c.Code, code)
	data.Set(c.Name, name)
	data.Set(c.Password, pw)
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		return errors.New(fmt.Sprintf("Creating of user %q failed", ph))
	}
	return nil
}

func initiateSignUp(ph string) error {
	data := url.Values{}
	data.Set(c.Phone, ph)
	req, _ := http.NewRequest(http.MethodPost, "/initiateSignUp", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res := executeRequest(req)
	if res.Code != http.StatusOK {
		return errors.New(fmt.Sprintf("Initiating of user signup %q failed", ph))
	}
	return nil
}

func testPhone(s string) string {
	mod := 10000007
	sum := 0
	for _, c := range s {
		sum = (sum*10001 + int(c)) % mod
	}
	return fmt.Sprintf("99%08d", sum)
}

func testName(s string) string {
	return s
}

func testPassword(s string) string {
	return fmt.Sprintf("123@%sA", s)
}

func afterTests() {
	dropDatabase()
	datastore.CloseMySql()
}

func dropDatabase() {
	// Delete the created database to clear the state
	db, err := sql.Open("mysql", c.DbUriBase)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := fmt.Sprintf(`
		DROP DATABASE IF EXISTS %s;`,
		c.TestMysqlDbName)

	if err := datastore.PrepareAndExec(query, db); err != nil {
		panic(err)
	}
}

func orderInIt(productId string, orderDate int64, price int, tax int, shippingCost int, addressId int, amount int, saleid int) types.Order {
	var order types.Order
	order.ProductId = productId
	order.OrderDate = orderDate
	order.Price = price
	order.Tax = tax
	order.ShippingCost = shippingCost
	order.AddressId = addressId
	order.Amount = amount
	order.SaleId = saleid
	return order
}
