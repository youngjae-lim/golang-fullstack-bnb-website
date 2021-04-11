package repository

import "github.com/youngjae-lim/golang-fullstack-bnb-website/internal/models"

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) error
}