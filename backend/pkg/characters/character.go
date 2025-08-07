package characters

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	database "fucku/internal/database"
	users "fucku/internal/users"
	utils "fucku/internal/utils"

	"github.com/jackc/pgx/v5"
)

type Character struct {
	Id        string     `json:"id"`
	UserId    string     `json:"user_id"`
	Name      string     `json:"name"`
	Inventory *Inventory `json:"inventory,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Inventory struct {
	Id          string         `json:"id"`
	CharacterId string         `json:"character_id"`
	Slots       int            `json:"slots"`
	Rows        []InventoryRow `json:"rows"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type InventoryRow struct {
	Id          string    `json:"id"`
	InventoryId string    `json:"inventory_id"`
	ItemId      string    `json:"item_id"`
	ItemType    string    `json:"item_type"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreateCharacterHandler(db *database.Database, logger *slog.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := users.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusBadRequest)
			return
		}

		var c Character
		error := utils.DecodeJSONBody(w, r, &c)
		if error != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Check if name is unique
		var name string
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		row := db.DBPool.QueryRow(ctx, `SELECT name FROM characters WHERE name = $1`, c.Name)
		if row.Scan(&name) != pgx.ErrNoRows {
			http.Error(w, "name is already taken", http.StatusBadRequest)
			return
		}

		_, err := db.DBPool.Exec(ctx, `INSERT INTO characters (user_id, name) VALUES ($1, $2)`, user.Id, c.Name)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Error("failed to create character", "error", err)
			return
		}

		created, err := GetCharacter(&user, db)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Error("failed to query character", "error", err)
			return
		}

		logger.Info("created character", "character", created)

		// Create character inventory
		_, err = db.DBPool.Exec(ctx, `INSERT INTO inventories (character_id) VALUES ($1)`, created.Id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Error("failed to create character inventory", "error", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(created)
		if err != nil {
			http.Error(w, "failed to encode character", http.StatusInternalServerError)
			return
		}
	})
}

func GetCharacterHandler(db *database.Database) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := users.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusBadRequest)
			return
		}

		character, err := GetCharacter(&user, db)
		if err != nil {
			http.Error(w, "character not found", http.StatusBadRequest)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(character)
		if err != nil {
			http.Error(w, "failed to encode character", http.StatusInternalServerError)
			return
		}
	})
}

func GetCharacter(u *users.User, db *database.Database) (*Character, error) {
	var c Character
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	row := db.DBPool.QueryRow(ctx,
		`SELECT id, user_id, name, created_at, updated_at FROM characters WHERE user_id = $1`,
		u.Id)

	if err := row.Scan(&c.Id, &c.UserId, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}

	// Remove empty inventory struct
	c.unsetInventory()

	return &c, nil
}

func CreateInventory(c *Character) error {
	return nil
}

func (c *Character) unsetInventory() {
	c.Inventory = nil
}
