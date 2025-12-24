package message

// Success messages
const (
	MsgRegisterSuccess = "User registered successfully"
	MsgLoginSuccess    = "Login successful"
	MsgTokenRefreshed  = "Token refreshed successfully"
)

// Error messages
const (
	ErrInvalidRequest     = "Invalid request payload"
	ErrEmailAlreadyExists = "Email already registered"
	ErrInvalidCredentials = "Invalid email or password"
	ErrUserNotFound       = "User not found"
	ErrInternalServer     = "Internal server error"
	ErrUnauthorized       = "Unauthorized access"
	ErrInvalidToken       = "Invalid or expired token"
	ErrHashPassword       = "Failed to hash password"
	ErrCreateUser         = "Failed to create user"
)
