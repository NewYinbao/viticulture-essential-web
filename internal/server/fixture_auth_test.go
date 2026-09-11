package server

// Fixture sessions represent already-enrolled players. Legacy migration is
// exercised separately with unprotected sessions in password_test.go.
var fixturePassword = makePassword("test-password-123")

func setFixtureSession(a *App, token, code, playerID string) {
	a.Store.Sessions[token] = Session{Code: code, PlayerID: playerID}
	a.Store.Passwords[playerID] = fixturePassword
}
