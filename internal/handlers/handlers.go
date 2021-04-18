package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	//"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/config"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/driver"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/forms"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/helpers"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/models"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/render"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/repository"
	"github.com/youngjae-lim/golang-fullstack-bnb-website/internal/repository/dbrepo"
)

// Repo is the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new repository for Testing only
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// send the data to the template
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Generals renders the Generals-Qaurters Room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the Majors-Suite Room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// Availability renders the search-availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability handles request for room availability
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get user-searched arrival date and departure date
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	// date format conversion
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get available rooms searched by date range
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// if there is no available rooms, show notie error message
	if len(rooms) == 0 { // no availability
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// ChooseRoom displays a list of available rooms that are linked to the make-reservation page
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// Get a param "id" from url
	// roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	// if err != nil {
	// 	helpers.ServerError(w, err)
	// 	m.App.Session.Put(r.Context(), "error", "missing url parameter")
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// 	return
	// }

	// changed how to get a param id (above) to this, so we can test it more easily
	// split the URL up by /, and grab the 3rd element (i.e., id if it exists)
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

// Reservation renders the make-reservation page and displays form with reservation details info
// Note that the make-reservation page is rendered upon a user making a request for room availability searched by an arrival data and a departure date
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// Get a session data for reservation
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get a reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	// Get a room name by id
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find a room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	res.Room.RoomName = room.RoomName

	// Store the room name to the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// cast time.Time date format to YYYY-MM-DD
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	// Get a session data (StartDate, EndDate, RoomID, RoomName) for reservation
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get a reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse the form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	// Put the updated reservation data back into the session
	m.App.Session.Put(r.Context(), "reservation", res)

	form := forms.New(r.PostForm)

	// Checks for form validation
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = res
		http.Error(w, "form data invalid!", http.StatusSeeOther)
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(res)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     res.StartDate,
		EndDate:       res.EndDate,
		RoomID:        res.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1, // TODO change this. hard-coded for now.
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//*******************************************
	// send email notification to guest
	//*******************************************
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s, <br>
		This is to confirm your reservation of %s from %s to %s.
	`, res.Room.RoomName, res.FirstName, res.StartDate.Format("2006-01-02"), res.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       res.Email,
		From:     "me@here.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	//*******************************************
	// send email notification to property owner
	//*******************************************
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Notification</strong><br>
		A reservation has been made for %s from %s to %s.
	`, res.Room.RoomName, res.StartDate.Format("2006-01-02"), res.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:       "me@here.com",
		From:     "me@here.com",
		Subject:  "Reservation Notification",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

// ReservationSummary displays reservation summary upon reservation
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Remove the reservation data stored in the session
	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = res

	// cast time.Time date format to YYYY-MM-DD
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for room availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	// parse request body
	err := r.ParseForm()
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	// Get arrival and departure dates from modal form
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	// convert date format to time.Time
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Arrival date invalid",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Departure date invalid",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Room ID invalid",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "     ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// BookRoom takes URL query parameters, builds a sessional reservation variable, and takes a user to make a reservation
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// Get query values from url
	// example: /book-room?id=1&s=2021-04-26&e=2021-04-29
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	var res models.Reservation

	// convert date format to time.Time
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get a room name by id
	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't get room from db!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// ShowLogin renders a login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging a user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully!!") // TODO: this message keeps popping up whenever we go to the previous page by hitting a back button while being logged in. Please fix this!
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// AdminDashboard shows dashboard content in the admin panel
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminNewReservations shows all new reservations in the admin panel
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations shows all new reservations in the admin panel
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminReservationDetail shows a reservation detail in the admin panel
func (m *Repository) AdminReservationDetail(w http.ResponseWriter, r *http.Request) {
	// Get {src} and {id} from URI strings in the Edit link
	uriStrings := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(uriStrings[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := uriStrings[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["year"] = year
	stringMap["month"] = month

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "admin-reservation-detail.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostReservationDetail...
func (m *Repository) AdminPostReservationDetail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Get {src} and {id} from URI strings in the Edit link
	uriStrings := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(uriStrings[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := uriStrings[3] // new or all

	stringMap := make(map[string]string)
	stringMap["src"] = src

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	// TODO: add validation
	// form := forms.New(r.PostForm)
	// form.Required("email")
	// form.IsEmail("email")
	// if !form.Valid() {
	// 	data := make(map[string]interface{})
	// 	data["reservation"] = reservation

	// 	render.Template(w, r, "admin-reservation-detail.page.tmpl", &models.TemplateData{
	// 		StringMap: stringMap,
	// 		Data:      data,
	// 		Form:      form,
	// 	})
	// }

	err = m.DB.UpdateReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.Form.Get("year")
	month := r.Form.Get("month")

	m.App.Session.Put(r.Context(), "flash", "Changes saved!")
	if year == "" { // saved through Reservations (all or new) menu
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else { // saved through Reservation Calendar menu
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// AdminReservationsCalendar shows reservation calendar in the admin panel
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified for now
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1) // add 1 month and then subtracts 1 day

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		// create maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		// initialize maps with zero values, which mean that there are no reservations or blocks
		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		// get all the restrictions for the current room for the current month
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		// for each room, mark dates as reservation or restriction
		for _, y := range restrictions { // for each restriction
			if y.ReservationID > 0 {
				// it's a reservation
				for d := y.StartDate; !d.Equal(y.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				// it's a block by a property owner
				// assume that a block by a property owner is done for each day through the admin panel.
				d := y.StartDate
				blockMap[d.Format("2006-01-2")] = y.ID

			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		IntMap:    intMap,
		Data:      data,
	})
}

// AdminPostReservationsCalendar handles post of reservation calendar
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// get values from hidden inputs inside form to be used to redirect a page after submitting a form
	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	// process blocks
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	// handle deleting blocks
	for _, x := range rooms {
		// get the block map from the session. Loop through the entire map, if we have an entry in the map that does not exist in our posted data, and if the restriction id > 0, then it is a block we need to remove
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for key, value := range curMap {
			// ok will be false if the value is not in the map
			val, ok := curMap[key]
			if ok {
				if val > 0 {
					// example: remove_block_1_2021-05-1
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, key)) {
						// delete the restriction by id
						err := m.DB.DeleteBlockByID(value)
						if err != nil {
							log.Println(err)
							helpers.ServerError(w, err)
							return
						}
					}
				}
			}
		}
	}

	// handle adding new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			// example: add_block_1_2021-05-9
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			startDate, _ := time.Parse("2006-01-2", exploded[3])
			// insert a new block
			err := m.DB.InsertBlockForRoom(roomID, startDate)
			if err != nil {
				log.Println(err)
				helpers.ServerError(w, err)
				return
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		log.Println(err)
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	log.Println(year)

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed!")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// AdminDeleteReservation deletes a reservation
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	err := m.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(), "flash", "Reservation successfully deleted!")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}
