package validate

import (
	"errors"
	"regexp"
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"strconv"
	"strings"
)

var Re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
var PhRe = regexp.MustCompile("^[789]\\d{9}$")

func Feed(lastSync, mascotId, flag string) (int64, int, int, error) {
	// lastSync should be valid integer
	ls, err := strconv.ParseInt(lastSync, 10, 64)
	if err != nil {
		return 0, 0, 0, errors.New("lastSync is not a valid Integer")
	}

	// mascotId should be valid integer
	ms, err := strconv.Atoi(mascotId)
	if err != nil {
		return 0, 0, 0, errors.New("mascotId is not a valid Integer")
	}

	// flag should be valid integer
	f, err := strconv.Atoi(flag)
	if err != nil {
		return 0, 0, 0, errors.New("flag is not a valid Integer")
	}

	// flag can take limited values
	if (f != c.TopPost) && (f != c.PostAfter) && (f != c.PostBefore) {
		return 0, 0, 0, errors.New("Invalid value for flag")
	}

	return ls, ms, f, nil
}

func Payment(orderId string) (int, error) {
	OrderId, err := strconv.Atoi(orderId)
	if err != nil {
		return 0, errors.New("orderId is not a valid Integer")
	}
	return OrderId, nil
}

func PaymentResponse(txid, amount, phone, email string) (int, int, int, string, error) {
	tId, err := strconv.Atoi(txid)
	if err != nil {
		return 0, 0, 0, "", errors.New("Invalid Transaction Id")
	}
	amt, err := strconv.Atoi(amount)
	if err != nil {
		return 0, 0, 0, "", errors.New("Invalid Amount")
	}
	validatePhone := ValidPhoneNumber(phone)
	var phn int
	if validatePhone {
		phn, _ = strconv.Atoi(phone)
	} else {
		return 0, 0, 0, "", errors.New("Invalid Phone Number")
	}
	validateEmail := validEmail(email)
	if validateEmail {
		return tId, amt, phn, email, nil
	} else {
		return 0, 0, 0, "", errors.New("Invalid Email")
	}

}

func CreatePost(cardType, src, dpSrc, title,
	desc, buttonText, url, icon, gradientStart,
	gradientEnd string, childPosts []string) (types.Post, error) {

	var p types.Post

	ctn, err := strconv.Atoi(cardType)
	if err != nil {
		return p, errors.New("Invalid cardType")
	}

	// valid card type?
	validCardType := false
	for _, x := range c.CardTypes {
		if ctn == x {
			validCardType = true
			break
		}
	}

	if !validCardType {
		return p, errors.New("Invalid cardType value")
	}

	if title == "" {
		return p, errors.New("Title cannot be empty")
	}

	datacard := false
	if ctn == c.CardTypeImage ||
		ctn == c.CardTypeArticle ||
		ctn == c.CardTypeGif {
		datacard = true
	}

	if datacard && dpSrc == "" {
		return p, errors.New("DpSrc cannot be empty for datacard")
	}

	if datacard && src == "" {
		return p, errors.New("Src cannot be empty for datacard")
	}

	if ctn == c.CardTypeArticle {
		if buttonText == "" {
			return p, errors.New("ButtonText cannot be empty for articleCard")
		}
		if url == "" {
			return p, errors.New("Url cannot be empty for articleCard")
		}
	}

	if ctn == c.CardTypeList {
		if len(childPosts) == 0 {
			return p, errors.New("ChildPosts cannot be empty for ListCard")
		}
		if icon == "" || gradientStart == "" || gradientEnd == "" {
			return p, errors.New("Icon, GradientStart, GradientEnd cannot be empty for ListCard")
		}
	}

	p.CardType = ctn
	p.Title = title
	p.DpSrc = dpSrc
	p.Src = src
	p.Description = desc
	p.ButtonText = buttonText
	p.Url = url
	p.Icon = icon
	p.GradientStart = gradientStart
	p.GradientEnd = gradientEnd
	p.ChildPosts = childPosts

	return p, nil
}

func ValidPhoneNumber(p string) bool {
	return PhRe.MatchString(p)
}

func validEmail(em string) bool {
	em = strings.ToLower(em)
	return Re.MatchString(em)
}

func InitiateSignUp(phone string) (string, error) {

	if ValidPhoneNumber(phone) {
		return phone, nil
	}

	return "", errors.New("Invalid phone number.")
}

func SignUp(phone, code, name, password string) (string, string, string, string, error) {
	if !ValidPhoneNumber(phone) {
		return "", "", "", "", errors.New("Invalid phone number.")
	}
	if code == "" {
		return "", "", "", "", errors.New("OTP code cannot be empty")
	}
	if name == "" {
		return "", "", "", "", errors.New("Name cannot be empty")
	}
	if len(password) < 8 {
		return "", "", "", "", errors.New("Password should be atleast 8 chars")
	}
	return phone, code, name, password, nil
}

func Login(phone, password string) (string, string, error) {
	if !ValidPhoneNumber(phone) {
		return "", "", errors.New("Invalid phone number.")
	}
	if len(password) < 8 {
		return "", "", errors.New("Password should be atleast 8 chars")
	}
	return phone, password, nil
}
