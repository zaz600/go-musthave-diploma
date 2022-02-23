package random

// SessionID случайный идентификатор пользовательской сессии
func SessionID() string {
	return String(32)
}
