package types

import "database/sql"
import "gopkg.in/mgo.v2/bson"

type User struct {
	Id                 int            `json:"-"`
	Email              string         `json:",omitempty"`
	Password           string         `json:"-"`
	Gender             sql.NullString `json:",omitempty"`
	FirstName          sql.NullString `json:",omitempty"`
	LastName           sql.NullString `json:",omitempty"`
	Phone              sql.NullString `json:",omitempty"`
	TimeOfCreation     int64          `json:"-"`
	Verified           int            `json:",omitempty"`
	Code               string         `json:"-"`
	Token              sql.NullString `json:"-"` // This ID is the Firebase registration token of each client
	ResetPasswordToken sql.NullString `json:"-"`
}
type Role struct {
	Id   int
	Name string
}

type PostLink struct {
	TimeOfCreation int64
	PostId         string
	MascotId       int
}

type Post struct {
	Id             bson.ObjectId `bson:"_id"`
	TimeOfCreation int64
	CardType       int
	Src            string
	DpSrc          string
	Title          string
	Description    string
	Url            string
	ButtonText     string
	ChildPosts     []string
	TimeOfLink     int64  `bson:"-"` // Value useful for frontend. Never in mongodb
	ChildPostsJson string `bson:"-"` // Value useful for frontend. Never in mongodb
	GradientStart  string
	GradientEnd    string
	Icon           string
}

type Posts []string
type QueueMap map[int]Posts

type Product struct {
	Id             bson.ObjectId `bson:"_id"`
	Sku            string
	Title          string
	Brand          string
	Quantity       int
	Description    string
	UnitPrice      int
	Summary        string
	Image          string
	ThumbNail      string
	Color          string
	Size           string
	TimeOfCreation int64
}

type Order struct {
	Id             int
	ProductId      string
	ProductTitle   string
	ProductThumb   string
	UserId         int `json:"-"`
	OrderDate      int64
	Price          int
	Tax            int
	ShippingCost   int
	Amount         int
	TransId        int
	TransStatus    string
	SaleId         int
	AddressId      int
	ShippingId     int
	ShippingStatus string
	TrackingId     string
	TimeOfCreation int64
}

type Shipping struct {
	Id             int
	OrderId        int
	UserId         int `json:"-"`
	TrackingId     string
	AddressId      int
	ShippingStatus string
	TimeOfCreation int64
}

type Address struct {
	Id             int
	UserId         int `json:"-"`
	Address        string
	AddressType    string
	City           string
	State          string
	PostalCode     int
	Phone          string
	TimeOfCreation int64
}

type Sale struct {
	Id             int
	Title          string
	Brand          string
	ProductSku     string
	Description    string
	ThumbNail      string
	StockUnits     int
	SaleStartTime  int64
	SaleEndTime    int64
	TimeOfCreation int64
}

type SalesList struct {
	Data []Sale
}

type OrdersList struct {
	Data []Order
}

type AddressList struct {
	Data []Address
}

type Transaction struct {
	Id             int
	Amount         int
	OrderId        int
	Phone          int
	TimeOfCreation int64
	ProductInfo    string
	Email          string
	PaymentMethod  string
	PaymentId      string
	PaymentStatus  string
	FirstName      string
	Hash           string
}

type TransactionResponse struct {
	TransId        int
	PaymentId      string
	PaymentStatus  string
	PaymentMessage string
}

type StatusResponse struct {
	TimeToStart int64
	StockLeft   int
}

// String hashSequence = key|txnid|amount|productinfo|firstname|email|udf1|udf2|udf3|udf4|udf5||||||salt;
type HashResponse struct {
	Hash        string
	TxnId       int
	Amount      int
	ProductInfo string
	FirstName   string
	Email       string
	Key         string
	Phone       int
	Surl        string
	Furl        string
	//	MerchantId  string
}

type InitiateSignUpResponse struct {
	Sum     int
	Product int
}
