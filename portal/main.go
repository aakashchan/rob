package main

import (
	//	"database/sql"
	//"errors"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	//"golang.org/x/crypto/bcrypt"
	//mgo "gopkg.in/mgo.v2"
	"html/template"
	"io"
	"net/http"
	"os"
	c "rob/lib/common/constants"
	//"rob/lib/common/httperr"
	//lh "rob/lib/common/loghelper"
	"rob/lib/common/types"
	//	"rob/lib/datastore"
	mw "rob/lib/middleware"
	//	"rob/lib/queue"
	//"rob/lib/session"
	"bytes"
	"io/ioutil"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type gzreadCloser struct {
	*gzip.Reader
	io.Closer
}

func (gz gzreadCloser) Close() error {
	return gz.Closer.Close()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		if k == "Content-Length" {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "tmpl/login.html")
		return
	}

	hc := http.Client{}
	em := r.FormValue(c.Email)
	pw := r.FormValue(c.Password)
	data := url.Values{}
	data.Set(c.Email, em)
	data.Set(c.Password, pw)
	req, err := http.NewRequest(http.MethodPost, c.LocalHost+"/login", bytes.NewBufferString(data.Encode()))
	if err != nil {
		http.Redirect(w, r, "/error", 302)
	}
	//	rDump, err := httputil.DumpRequest(r, true)
	//	log.Debug(string(rDump))
	//log.Debug(fmt.Sprintf("R header : %v", r.Header))
	//log.Debug(fmt.Sprintf("R Body: %v", r.Body))

	copyHeader(req.Header, r.Header)
	//reqDump, err := httputil.DumpRequest(req, true)
	//log.Debug(string(reqDump))

	log.Debug(fmt.Sprintf("Req header : %v", req.Header))
	log.Debug(fmt.Sprintf("Req Body: %v", req.Body))
	res, err := hc.Do(req)
	if err != nil {
		http.Redirect(w, r, "/error", 302)
	}
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	bodyString := string(bodyBytes)
	log.Debug(fmt.Sprintf("Req header : %v", r.Header))
	log.Debug(fmt.Sprintf("Req Body: %s", bodyString))
	if res.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/error", 302)
	}
	defer res.Body.Close()
	log.Debug(fmt.Sprintf("Res header : %v", res.Header))
	log.Debug(fmt.Sprintf("Res Body: %v", res.Body))
	// headers
	//resDump, err := httputil.DumpResponse(res, true)
	//log.Debug(string(resDump))

	for name, values := range res.Header {
		if name == "Content-Length" {
			continue
		} else if name == "Content-Type" {
			w.Header()[name] = []string{"text/html"}
			continue
		}
		w.Header()[name] = values
	}

	// status (must come after setting headers and before copying body)

	// body

	//io.Copy(w, res.Body)
	//io.Copy(w, bytes.NewBuffer([]byte("Testing\n")))
	io.WriteString(w, "<html><body><script>window.location.replace('/list');</script></body></html>")
	log.Debug(fmt.Sprintf("Write header : %v", w.Header()))
	log.Printf("Write Body: %#v", w)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Logout Handler")
	hc := http.Client{}
	log.Debug(fmt.Sprintf(" Logout R header : %v", r.Header))
	log.Debug(fmt.Sprintf("Logout R Body: %v", r.Body))
	req, err := http.NewRequest(http.MethodGet, c.LocalHost+"/logout", nil)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	copyHeader(req.Header, r.Header)
	log.Debug(fmt.Sprintf("Logout Req: %s %s\n", req.Host, req.URL.Path))
	log.Debug(fmt.Sprintf("Logout Req header : %v", req.Header))
	log.Debug(fmt.Sprintf("Logout Req Body: %v", req.Body))

	res, err := hc.Do(req)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	if res.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/error", 302)
	}
	for name, values := range res.Header {
		if name == "Content-Length" {
			continue
		} else if name == "Content-Type" {
			w.Header()[name] = []string{"text/html"}
			continue
		}
		w.Header()[name] = values
	}

	// status (must come after setting headers and before copying body)

	// body

	//io.Copy(w, res.Body)
	//io.Copy(w, bytes.NewBuffer([]byte("Testing\n")))
	io.WriteString(w, "<html><body><script>window.location.replace('/login');</script></body></html>")
	log.Debug(fmt.Sprintf("Write header : %v", w.Header()))
	log.Printf("Write Body: %#v", w)

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	hc := http.Client{}
	req, err := http.NewRequest(http.MethodGet, c.LocalHost+"/posts", nil)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}

	copyHeader(req.Header, r.Header)

	res, err := hc.Do(req)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	if res.StatusCode != http.StatusOK {
		log.Debug(fmt.Sprintf("Status Code : %d", res.StatusCode))
		http.Redirect(w, r, "/error", 302)
	}

	defer res.Body.Close()
	resDump, err := httputil.DumpResponse(res, true)
	log.Debug(string(resDump))
	log.Debug(fmt.Sprintf("List Res header : %v", res.Header))
	log.Debug(fmt.Sprintf("List Res Body: %v", res.Body))
	var result []types.Post
	if res.Header.Get("Content-Encoding") == "gzip" {
		res.Header.Del("Content-Length")
		zr, err := gzip.NewReader(res.Body)
		if err != nil {
			log.Error(err.Error())
		}
		res.Body = gzreadCloser{zr, res.Body}
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Error(err)
	}

	log.Debug(fmt.Sprintf("Result : %#v", result))

	//fmt.Sprintf("data : %v", result)
	t, err := template.New("list.html").ParseFiles("tmpl/list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	hc := http.Client{}
	var p types.Post
	r.ParseMultipartForm(32 << 20)
	//	p.Src = r.FormFile(c.Src)
	file, handler, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/error", 302)
	}
	defer file.Close()
	//fmt.Fprintf(w, "%v", handler.Header)
	suff := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	f, err := os.OpenFile("./uploads/"+suff+"_"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/error", 302)
	}
	defer f.Close()
	io.Copy(f, file)
	p.Src = "https://cdn.twiq.in/uploads/" + suff + "_" + handler.Filename
	p.DpSrc = r.FormValue(c.DpSrc)
	p.Title = r.FormValue(c.Title)
	p.Description = r.FormValue(c.Description)
	p.ButtonText = r.FormValue(c.ButtonText)
	p.Url = r.FormValue(c.Url)
	data := url.Values{}
	ct := r.FormValue(c.CardType)
	_, err = strconv.Atoi(ct)
	if err != nil {
		log.Error("Invalid Request Parameter cardType")
		http.Redirect(w, r, "error", 302)
	}
	data.Set(c.CardType, ct)
	data.Set(c.DpSrc, p.DpSrc)
	data.Set(c.Title, p.Title)
	data.Set(c.Src, p.Src)
	data.Set(c.Description, p.Description)
	data.Set(c.ButtonText, p.ButtonText)
	data.Set(c.Url, p.Url)
	req, err := http.NewRequest(http.MethodPost, c.LocalHost+"/post", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err)
	}
	copyHeader(req.Header, r.Header)
	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	log.Debug(fmt.Sprintf("Post Req: %s %s\n", req.Host, req.URL.Path))
	log.Debug(fmt.Sprintf("Post Req header : %v", req.Header))
	log.Debug(fmt.Sprintf("Post Req Body: %v", req.Body))

	res, err := hc.Do(req)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}

	if res.StatusCode != http.StatusOK {
		log.Error(fmt.Sprintf("Create Post request failed. Title=%s. Status=%d", p.Title, res.StatusCode))
		http.Redirect(w, r, "/error", 302)
	}
	defer res.Body.Close()
	var postId string
	err = json.NewDecoder(res.Body).Decode(&postId)
	if err != nil {
		log.Error(err)
	}
	hc = http.Client{}
	data = url.Values{}
	data.Set(c.PostId, postId)
	data.Set(c.MascotId, "1")
	req, err = http.NewRequest(http.MethodPost, c.LocalHost+"/postlink", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	copyHeader(req.Header, r.Header)
	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	res, err = hc.Do(req)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	if res.StatusCode != http.StatusOK {
		log.Error("Could not link the post %s to mascotId 1", postId)
		http.Redirect(w, r, "/error", 302)
	}

	//resDump, err := httputil.DumpResponse(res, true)
	//log.Debug(string(resDump))
	http.Redirect(w, r, "/success", 302)

}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	hc := http.Client{}
	postId := r.FormValue(c.PostId)
	data := url.Values{}
	data.Set(c.PostId, postId)
	req, err := http.NewRequest(http.MethodPost, c.LocalHost+"/deletePost", strings.NewReader(data.Encode()))
	if err != nil {
		log.Error(err)
	}
	copyHeader(req.Header, r.Header)

	log.Debug(fmt.Sprintf("Post Req: %s %s\n", req.Host, req.URL.Path))
	log.Debug(fmt.Sprintf("Post Req header : %v", req.Header))
	log.Debug(fmt.Sprintf("Post Req Body: %v", req.Body))

	res, err := hc.Do(req)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}

	if res.StatusCode != http.StatusOK {
		log.Error(fmt.Sprintf("Delete Post request failed. PostId=%s. Status=%d", postId, res.StatusCode))
		http.Redirect(w, r, "/error", 302)
	}
	defer res.Body.Close()
	resDump, err := httputil.DumpResponse(res, true)
	log.Debug(string(resDump))
	http.Redirect(w, r, "/success", 302)

}

func postlinkHandler(w http.ResponseWriter, r *http.Request) {
	hc := http.Client{}
	postId := r.FormValue(c.PostId)
	mascotId := r.FormValue(c.MascotId)
	data := url.Values{}
	data.Set(c.PostId, postId)
	data.Set(c.MascotId, mascotId)
	req, err := http.NewRequest(http.MethodPost, c.LocalHost+"/postlink", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	copyHeader(req.Header, r.Header)
	res, err := hc.Do(req)
	if err != nil {
		log.Error(err.Error())
		http.Redirect(w, r, "/error", 302)
	}
	if res.StatusCode != http.StatusOK {
		log.Error("Could not link the post %s to mascotId %s", postId, mascotId)
		http.Redirect(w, r, "/error", 302)
	}

	http.Redirect(w, r, "/success", 302)
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "tmpl/success.html")
	}

}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "tmpl/error.html")
	}

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
	initLogging()
	r := mux.NewRouter()

	r.Handle("/login",
		alice.New(mw.NoAuth).
			ThenFunc(loginHandler)).
		Methods("GET", "POST")

	r.Handle("/logout",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(logoutHandler)).
		Methods("GET")

	r.Handle("/list",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(listHandler)).
		Methods("GET")

	r.Handle("/create",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(createHandler)).
		Methods("POST")

	r.Handle("/delete",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(deleteHandler)).
		Methods("POST")

	r.Handle("/postlink",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(postlinkHandler)).
		Methods("POST")

	r.Handle("/success",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(successHandler)).
		Methods("GET")

	r.Handle("/error",
		alice.New(mw.Auth).
			Append(mw.CheckAccess(c.AdminRole)).
			ThenFunc(errorHandler)).
		Methods("GET")

	s := http.StripPrefix("/statics/", http.FileServer(http.Dir("./statics/")))
	r.PathPrefix("/statics/").Handler(s)
	return mw.Common(r)
}

func main() {

	http.Handle("/", getRouter())

	log.Info("Server running on port 8898")

	err := http.ListenAndServe(":8898", nil)
	if err != nil {
		log.Error(err)
	}
}
