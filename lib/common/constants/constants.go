package constants

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// App wide we are going to use a common naming mechanism
// Everything will be starting with capital letter.
// That is because in go capital letter is public, and so in many structs
// we will have to use it as capital letter
// We will use FirstLetterCapital style and not under_score style
// Exceptions are for things like dbname, urls etc
// As long as it's readable and understandable we will use shortest names possible
// For example "TransId" instead of "TransactionId"
var (
	Email              = "Email"
	Password           = "Password"
	Gender             = "Gender"
	FirstName          = "FirstName"
	LastName           = "LastName"
	IsActive           = "IsActive"
	RoleId             = "RoleId"
	User               = "User"
	Id                 = "Id"
	DbUriBase          = "root:ubuntu@/"
	DbName             = "disha"
	DbUri              = fmt.Sprintf("%s%s", DbUriBase, DbName)
	TimeOfCreation     = "TimeOfCreation"
	Phone              = "Phone"
	UserId             = "UserId"
	Name               = "Name"
	NumOfPosts         = 150
	Verified           = "Verified"
	Code               = "Code"
	Confirm            = 1
	Token              = "Token"
	ResetPasswordToken = "ResetPasswordToken"
	NewPassword        = "NewPassword"
	NewPasswordRepeat  = "NewPasswordRepeat"
)

var (
	CardType      = "CardType"
	ChildPosts    = "ChildPosts"
	DpSrc         = "DpSrc"
	ButtonText    = "ButtonText"
	Url           = "Url"
	Icon          = "Icon"
	GradientStart = "GradientStart"
	GradientEnd   = "GradientEnd"
)

// Allowed card types
var (
	CardTypeImage   = 0
	CardTypeArticle = 1
	CardTypeGif     = 2
	CardTypeDateSep = 3
	CardTypeList    = 4
	// When updating this, update the below array
)

// All mysql database tables as an array
var CardTypes = []int{
	CardTypeImage,
	CardTypeArticle,
	CardTypeGif,
	CardTypeDateSep,
	CardTypeList,
}

var (
	StatusNotVerified = "NotVerified"
	StatusVerified    = "Verified"
	StatusDeleted     = "Deleted"
)

var (
	Server            = "localhost:27017"
	Collection        = "posts"
	CounterCollection = "counters"
	ProductCollection = "products"
)

var (
	Title                    = "Title"
	Description              = "Description"
	Src                      = "Src"
	DefaultMascotId          = 1
	DefaultMascotName        = "Default Mascot"
	DefaultMascotDescription = "This is a default mascot"
)

var (
	PostId           = "PostId"
	MascotId         = "MascotId"
	LastSync         = "LastSync"
	DefaultTimestamp = -1
	Flag             = "Flag"
)

var (
	AdminRole  = 1
	UserRole   = 2
	WriterRole = 3
)
var (
	AdminRoleName  = "Admin"
	UserRoleName   = "UserRole"
	WriterRoleName = "WriterRole"
)

// All mysql database tables
var (
	UsersTable       = "Users"
	RolesTable       = "Roles"
	UserRoleTable    = "UserRole"
	MascotTable      = "Mascot"
	PostQueueTable   = "PostQueue"
	SaleTable        = "Sale"
	OrderTable       = "Orders"
	ShippingTable    = "Shipping"
	FeedbackTable    = "Feedback"
	AddressTable     = "Address"
	TransactionTable = "Transaction"
	UrlCacheTable    = "Url"
	// When updating this, update the below array
)

// All mysql database tables as an array
var MysqlTables = []string{
	UsersTable,
	RolesTable,
	UserRoleTable,
	MascotTable,
	PostQueueTable,
	SaleTable,
	OrderTable,
	ShippingTable,
	FeedbackTable,
	AddressTable,
	TransactionTable,
}

// Variables related to feedback
// To keep it simple and easy for now a lot of endpoints use just feedback table
// type column in the database will tell if it is related to "feedback" or
// "contact us" or "complaint" or so on
var (
	Type      = "Type"
	Feedback  = "Feedback"
	ContactUs = "ContactUs"
)

// Varibles related to product
var (
	Sku       = "Sku"
	Brand     = "Brand"
	Quantity  = "Quantity"
	UnitPrice = "UnitPrice"
	Summary   = "Summary"
	Image     = "Image"
	ThumbNail = "ThumbNail"
	Stock     = "Stock"
	Color     = "Color"
	Size      = "Size"
)

// Variables related to Sale
var (
	ProductSku    = "ProductSku"
	StockUnits    = "StockUnits"
	SaleStartTime = "SaleStartTime"
	SaleEndTime   = "SaleEndTime"
)

// Variables related to configuration
var (
	EnvDebug          = "DISHA_DEBUG"
	CredsBase         = ".creds"
	MysqlCredsFile    = CredsBase + "/.mysql"
	AWSCredsFile      = CredsBase + "/.aws"
	CcavenueCredsFile = CredsBase + "/.ccavenue"
	CcavenuePemFile   = CredsBase + "/.ccavenue.pem"
	PayUCredsFile     = CredsBase + "/.payu"
)

// Variables related to PayU
var (
	PayUKey     = ""
	PayUSalt    = ""
	Surl        = "https://twiq.in/api/payment-success"
	Furl        = "https://twiq.in/api/payment-failure"
	Hash        = "Hash"
	ProductInfo = "ProductInfo"
)

// Variables related to aws ses & mailing module
var (
	AWSSESProfile = "ses"
	EmailInfo     = "info@twiq.in"
)

// Variables related to testing
var (
	TestMysqlUser   = "disha_test"
	TestMysqlPass   = "Disha@1test"
	TestMysqlDbName = "dishatest"
)

// Variables related to Order
var (
	ProductId      = "ProductId"
	ProductTitle   = "ProductTitle"
	ProductThumb   = "ProductThumb"
	OrderDate      = "OrderDate"
	Price          = "Price"
	Tax            = "Tax"
	ShippingCost   = "ShippingCost"
	Amount         = "Amount"
	TransId        = "TransId"
	TransStatus    = "TransStatus"
	SaleId         = "SaleId"
	ShippingId     = "ShippingId"
	ShippingStatus = "ShippingStatus"
	TrackingId     = "TrackingId"
	Uninitiated    = "Uninitiated"
	UninitiatedId  = -1
	OrderId        = "OrderId"
)

// Variables related to Address
var (
	AddressId   = "AddressId"
	Address     = "Address"
	AddressType = "AddressType"
	City        = "City"
	State       = "State"
	PostalCode  = "PostalCode"
)

// Variables related to Transaction
var (
	PaymentMethod = "PaymentMethod"
	PaymentId     = "PaymentId"
	PaymentStatus = "PaymentStatus"
	DefaultString = "Default"
	DefaultInt    = -1
)

// variable related to Payment Gateway
var (
	ApiKey    = "randomKeyForNow"
	ApiSecret = "randomSecretForNow"
)

// Variable related StatusResponse
var (
	TimeToStart = "TimeToStart"
	Stockleft   = "StockLeft"
)

// Misc

var (
	UnAuthorized = "UnAuthorized"
	LocalHost    = "http://localhost:9980"
	TwiqUrl      = "https://twiq.in/api"
	TopPost      = 1
	PostBefore   = 2
	PostAfter    = 3
)

var DE = false
var EV = true // Email verification ON/OFF
func SetMysqlCreds(user, pass, dbName string) {
	DbUriBase = fmt.Sprintf("%s:%s@/", user, pass)
	DbName = dbName
	DbUri = fmt.Sprintf("%s%s", DbUriBase, DbName)
}

func init() {
	// To start the binary in debug mode simply set environment variable.
	// 'export DEBUG_DISHA=TRUE'
	if os.Getenv(EnvDebug) == "TRUE" {
		log.Info("Started in global debug mode")
		DE = true
	}
	// Ability to override default database username/password using
	// credentials file. .mysql_creds should be present in current
	// running directory
	// Format for the creds file is username on line 1 and password on line 2
	// and both lines ending with '\n'

	file, err := os.Open(MysqlCredsFile)
	if err == nil {
		defer file.Close()
		log.Info("Overridding mysql username/password with cred file")
		r := bufio.NewReader(file)

		mysqlUser, err := r.ReadString('\n')
		if err != nil {
			panic(fmt.Sprintf("Error reading %s file: %v", MysqlCredsFile, err.Error()))
		}

		mysqlPass, err := r.ReadString('\n')
		if err != nil {
			panic(fmt.Sprintf("Error reading %s file: %v", MysqlCredsFile, err.Error()))
		}

		mysqlUser = strings.TrimSuffix(mysqlUser, "\n")
		mysqlPass = strings.TrimSuffix(mysqlPass, "\n")

		SetMysqlCreds(mysqlUser, mysqlPass, DbName)
	}

	file, err = os.Open(PayUCredsFile)
	if err == nil {
		defer file.Close()
		log.Info("Setting up PayU Creds with cred file")
		r := bufio.NewReader(file)

		PayUKey, err = r.ReadString('\n')
		if err != nil {
			panic(fmt.Sprintf("Error reading %s file: %v", PayUCredsFile, err.Error()))
		}

		PayUSalt, err = r.ReadString('\n')
		if err != nil {
			panic(fmt.Sprintf("Error reading %s file: %v", PayUCredsFile, err.Error()))
		}

		PayUKey = strings.TrimSuffix(PayUKey, "\n")
		PayUSalt = strings.TrimSuffix(PayUSalt, "\n")

	} else {
		panic("Error Reading file. PayU Creds was not set")
	}
}
