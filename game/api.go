package game

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mreliasen/swi-server/internal"
	"github.com/mreliasen/swi-server/internal/database/models"
	"github.com/mreliasen/swi-server/internal/logger"
)

type CheckNameBody struct {
	Name string
}

type CheckNameResponse struct {
	Taken bool `json:"available"`
	Valid bool `json:"valid"`
}

type RegistrationBody struct {
	Email    string
	Password string
}

type RegistrationResponse struct {
	Error   bool     `json:"error"`
	Message string   `json:"message"`
	Fields  []string `json:"fields"`
}

func (g *Game) CheckNameTaken(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Logger.Warn(err.Error())
		err := json.NewEncoder(w).Encode(CheckNameResponse{
			Taken: false,
			Valid: false,
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	body := CheckNameBody{}
	json.Unmarshal(reqBody, &body)

	if !internal.IsValidCharacterName(body.Name) {
		err := json.NewEncoder(w).Encode(CheckNameResponse{
			Taken: false,
			Valid: false,
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	row := g.DbConn.QueryRow("SELECT id FROM characters WHERE LOWER(name) = ? LIMIT 1", strings.ToLower(body.Name))
	if row == nil {
		err := json.NewEncoder(w).Encode(CheckNameResponse{
			Taken: false,
			Valid: true,
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	character := models.Character{}
	row.Scan(&character.Id)

	err = json.NewEncoder(w).Encode(CheckNameResponse{
		Taken: character.Id == 0,
		Valid: true,
	})
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
}

func (g *Game) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		err := json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "Invalid Details",
			Fields:  []string{},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	body := RegistrationBody{}
	json.Unmarshal(reqBody, &body)

	if !strings.Contains(body.Email, "@") {
		err := json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "Invalid email",
			Fields: []string{
				"email",
			},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	if len(body.Password) < 8 {
		err := json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "Password must be at least 8 character long",
			Fields: []string{
				"password",
			},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	row := g.DbConn.QueryRow("SELECT email FROM users WHERE email = ? LIMIT 1", body.Email)
	user := models.User{}
	row.Scan(&user.Email)

	if user.Email != "" {
		err := json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "That email is already in use",
			Fields: []string{
				"email",
			},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	passwordHash, err := internal.HashPassword(body.Password)
	if err != nil {
		logger.Logger.Warn(err.Error())
		err := json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "Failed to create account, try again later.",
			Fields:  []string{},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	result, err := g.DbConn.Exec("INSERT INTO users (email, password) VALUES (?, ?)", body.Email, passwordHash)
	if err != nil {
		logger.Logger.Warn(err.Error())
		err := json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "Failed to create account, try again later.",
			Fields:  []string{},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	if _, err := result.LastInsertId(); err != nil {
		logger.Logger.Warn(err.Error())
		err = json.NewEncoder(w).Encode(RegistrationResponse{
			Error:   true,
			Message: "Failed to create account, try again later.",
			Fields:  []string{},
		})
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
		return
	}

	err = json.NewEncoder(w).Encode(RegistrationResponse{
		Error:   false,
		Message: "Account created! ",
	})
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
}
