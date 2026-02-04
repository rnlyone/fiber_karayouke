package models

import "time"

// Karaoke domain models representing the collaborative playlist schema.
type Subscription struct {
	ID            string `gorm:"column:id;primaryKey" json:"id"`
	PackageName   string `gorm:"column:package_name" json:"package_name"`
	PackageDetail []byte `gorm:"column:package_detail" json:"package_detail"`
	Price         string `gorm:"column:price" json:"price"`
	Visibility    bool   `gorm:"column:visibility" json:"visibility"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type UserSubscription struct {
	ID             string     `gorm:"column:id;primaryKey" json:"id"`
	SubscriptionID string     `gorm:"column:id_subs" json:"subscription_id"`
	UserID         string     `gorm:"column:id_user" json:"user_id"`
	PurchaseStamp  time.Time  `gorm:"column:purchase_stamp" json:"purchase_stamp"`
	ExpiredAt      *time.Time `gorm:"column:expired_at" json:"expired_at"`

	Subscription Subscription `gorm:"foreignKey:SubscriptionID;references:ID" json:"subscription"`
	User         User         `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (UserSubscription) TableName() string {
	return "user_subs"
}

type User struct {
	ID       string `gorm:"column:id;primaryKey" json:"id"`
	Name     string `gorm:"column:name" json:"name"`
	Username string `gorm:"column:username" json:"username"`
	Email    string `gorm:"column:email" json:"email"`
	Password string `gorm:"column:password" json:"-"`
	Credit   int    `gorm:"column:credit" json:"credit"`
}

func (User) TableName() string {
	return "users"
}

type Package struct {
	ID            string `gorm:"column:id;primaryKey" json:"id"`
	PackageName   string `gorm:"column:package_name" json:"package_name"`
	PackageDetail []byte `gorm:"column:package_detail" json:"package_detail"`
	Price         string `gorm:"column:price" json:"price"`
	CreditAmount  int    `gorm:"column:credit_amount" json:"credit_amount"`
	Visibility    bool   `gorm:"column:visibility" json:"visibility"`
}

func (Package) TableName() string {
	return "packages"
}

type Room struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	RoomKey     string     `gorm:"column:room_key;uniqueIndex" json:"room_key"`
	RoomCreator string     `gorm:"column:room_creator" json:"room_creator"`
	RoomName    string     `gorm:"column:room_name" json:"room_name"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ExpiredAt   *time.Time `gorm:"column:expired_at" json:"expired_at"`
	RoomMaster  string     `gorm:"column:room_master" json:"room_master"`
	Creator     User       `gorm:"foreignKey:RoomCreator;references:ID" json:"creator"`
	Master      User       `gorm:"foreignKey:RoomMaster;references:ID" json:"master"`
}

func (Room) TableName() string {
	return "rooms"
}

// IsExpired checks if the room has expired
func (r *Room) IsExpired() bool {
	if r.ExpiredAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiredAt)
}

type PurchaseLog struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	PackageID     string    `gorm:"column:id_package" json:"package_id"`
	UserID        string    `gorm:"column:id_user" json:"user_id"`
	PurchaseStamp time.Time `gorm:"column:purchase_stamp" json:"purchase_stamp"`
	Package       Package   `gorm:"foreignKey:PackageID;references:ID" json:"package"`
	User          User      `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (PurchaseLog) TableName() string {
	return "purchase_logs"
}

type Song struct {
	ID                 string     `gorm:"column:id;primaryKey" json:"id"`
	RoomID             string     `gorm:"column:room_id" json:"room_id"`
	Title              string     `gorm:"column:title" json:"title"`
	Artist             string     `gorm:"column:artist" json:"artist"`
	CoverURL           string     `gorm:"column:cover_url" json:"cover_url"`
	Duration           string     `gorm:"column:duration" json:"duration"`
	AddedAt            time.Time  `gorm:"column:added_at;autoCreateTime" json:"added_at"`
	PlayedAt           *time.Time `gorm:"column:played_at" json:"played_at"`
	RequestedByUserID  *string    `gorm:"column:requested_by_user" json:"requested_by_user"`
	RequestedByGuestID *string    `gorm:"column:requested_by_guest" json:"requested_by_guest"`
	Room               Room       `gorm:"foreignKey:RoomID;references:ID" json:"room"`
	RequestedByUser    *User      `gorm:"foreignKey:RequestedByUserID;references:ID" json:"requested_by_user_details"`
	RequestedByGuest   *Guest     `gorm:"foreignKey:RequestedByGuestID;references:ID" json:"requested_by_guest_details"`
}

func (Song) TableName() string {
	return "songs"
}

type Guest struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UserID    *string   `gorm:"column:id_user" json:"user_id"`
	RoomID    string    `gorm:"column:id_room" json:"room_id"`
	Room      Room      `gorm:"foreignKey:RoomID;references:ID" json:"room"`
	User      *User     `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (Guest) TableName() string {
	return "guests"
}

// SystemConfig stores global system configuration
type SystemConfig struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Key       string    `gorm:"column:key;uniqueIndex" json:"key"`
	Value     string    `gorm:"column:value" json:"value"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// Transaction represents a payment transaction
type Transaction struct {
	ID            string     `gorm:"column:id;primaryKey" json:"id"`
	UserID        string     `gorm:"column:user_id" json:"user_id"`
	PackageID     string     `gorm:"column:package_id" json:"package_id"`
	Amount        int64      `gorm:"column:amount" json:"amount"`
	Status        string     `gorm:"column:status" json:"status"` // pending, settlement, failed, expired, refunded
	PaymentMethod string     `gorm:"column:payment_method" json:"payment_method"`
	ExternalID    string     `gorm:"column:external_id" json:"external_id"` // For payment gateway reference
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	PaidAt        *time.Time `gorm:"column:paid_at" json:"paid_at"`
	User          User       `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Package       Package    `gorm:"foreignKey:PackageID;references:ID" json:"package"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// Transaction status constants
const (
	TransactionStatusPending    = "pending"
	TransactionStatusSettlement = "settlement"
	TransactionStatusFailed     = "failed"
	TransactionStatusExpired    = "expired"
	TransactionStatusRefunded   = "refunded"
)

// System config key constants
const (
	ConfigRoomMaxDuration  = "room_max_duration"  // in minutes
	ConfigRoomCreationCost = "room_creation_cost" // credits required
	ConfigDefaultCredits   = "default_credits"    // credits for new users
)

// CreditLog tracks credit changes
type CreditLog struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	UserID      string    `gorm:"column:user_id" json:"user_id"`
	Amount      int       `gorm:"column:amount" json:"amount"`             // positive for add, negative for deduct
	Balance     int       `gorm:"column:balance" json:"balance"`           // balance after transaction
	Type        string    `gorm:"column:type" json:"type"`                 // purchase, room_creation, admin_award, refund
	ReferenceID string    `gorm:"column:reference_id" json:"reference_id"` // transaction_id, room_id, etc.
	Description string    `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	User        User      `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (CreditLog) TableName() string {
	return "credit_logs"
}

// Credit log type constants
const (
	CreditTypeAdminAward   = "admin_award"
	CreditTypePurchase     = "purchase"
	CreditTypeRoomCreation = "room_creation"
	CreditTypeRefund       = "refund"
)

// Session stores user authentication sessions in the database
type Session struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Token     string    `gorm:"column:token;uniqueIndex" json:"token"`
	UserID    string    `gorm:"column:user_id;index" json:"user_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ExpiresAt time.Time `gorm:"column:expires_at;index" json:"expires_at"`
	User      User      `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (Session) TableName() string {
	return "sessions"
}

// TVToken stores TV device connection tokens
type TVToken struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Token     string    `gorm:"column:token;uniqueIndex" json:"token"`           // Full token for QR code
	ShortCode string    `gorm:"column:short_code;uniqueIndex" json:"short_code"` // 5-char code for manual entry
	RoomKey   string    `gorm:"column:room_key" json:"room_key"`                 // Connected room (empty until connected)
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ExpiresAt time.Time `gorm:"column:expires_at;index" json:"expires_at"`
}

func (TVToken) TableName() string {
	return "tv_tokens"
}
