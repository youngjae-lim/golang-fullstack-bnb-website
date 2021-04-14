package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/models"
)

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},

	// The belows are the ones that are using session. We need to test them separately!
	// {"reservation-summary", "/reservation-summary", "GET", []postData{}, http.StatusOK},
	// {"post-search-avail", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2021-08-20"},
	// 	{key: "end", value: "2021-08-22"},
	// }, http.StatusOK},
	// {"post-search-avail-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2021-08-20"},
	// 	{key: "end", value: "2021-08-22"},
	// }, http.StatusOK},
	// {"make-reservation-post", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "Youngjae"},
	// 	{key: "last_name", value: "Lim"},
	// 	{key: "email", value: "test@test.com"},
	// 	{key: "phone", value: "777-777-7777"},
	// }, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}

}

func TestRepository_Reservation(t *testing.T) {
	// test case where reservation is manually set in the session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, but expected %d", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code for missing session: got %d, but expected %d", rr.Code, http.StatusOK)
	}

	// test case where we can't get a room by id
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	reservation.RoomID = 100
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, but expected %d", rr.Code, http.StatusOK)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	// Get user-searched arrival date and departure date
	start := "2050-01-01"
	end := "2050-01-02"

	// date format conversion
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, start)
	endDate, _ := time.Parse(layout, end)
	
	//test case where reservation is manually set in the session
	reservation := models.Reservation{
		RoomID: 1,
		StartDate: startDate,
		EndDate: endDate,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	// set form data manually in the request body
	reqBody := "first_name=John"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Doe")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=test@test.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=666-666-6666")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, but expected %d", rr.Code, http.StatusSeeOther)
	}

	// test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, but expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case where reservation is not in session (reset everything)
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing session: got %d, but expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case for invalid form data: first_name only 1 character
	// set form data manually in the request body
	reqBody = "first_name=Y" // only 1 character
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Doe")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=test@test.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=666-666-6666")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr = httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid form data: got %d, but expected %d", rr.Code, http.StatusSeeOther)
	}

	// test case for non-existent RoomID in the session - can't insert data into reservations table in the database
	// set form data manually in the request body
	reqBody = "first_name=John" // only 1 character
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Doe")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=test@test.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=666-666-6666")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr = httptest.NewRecorder()
	reservation.RoomID = 2
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for non-existent RoomID: got %d, but expected %d", rr.Code, http.StatusTemporaryRedirect)
	}

	// test case for non-existent RestrictionID in the session - can't insert data into room_restrictions table in the database
	// set form data manually in the request body
	reqBody = "first_name=John" // only 1 character
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Doe")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=test@test.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=666-666-6666")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	rr = httptest.NewRecorder()
	reservation.RoomID = 1000
	session.Put(ctx, "reservation", reservation)

	handler = http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for non-existent RestrictionID: got %d, but expected %d", rr.Code, http.StatusTemporaryRedirect)
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
