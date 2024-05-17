package nest

// Account represents the data model for account.
type Account struct {
	ID       int
	Name     string
	Email    string
	Password string
}

// User represents the data model for a user.
type User struct {
	UserID   int      `json:"user_id"`
	Username string   `json:"username"`
	UserData UserData `json:"user_data"`
}

// UserData represents data owned by the user.
type UserData struct {
	Options map[string]interface{} `json:"options"`
	Data    Data                   // The fields of Data operate at depth level 3.
}

// Data represents a piece of data.
type Data struct {
	ID int
}
