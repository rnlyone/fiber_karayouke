package models

import "time"

// Karaoke domain models representing the collaborative playlist schema.

// SubscriptionPlan defines available subscription tiers (free, basic, pro, etc.)
type SubscriptionPlan struct {
	ID                  string `gorm:"column:id;primaryKey" json:"id"`
	PlanName            string `gorm:"column:plan_name" json:"plan_name"`
	PlanDetail          []byte `gorm:"column:plan_detail" json:"plan_detail"`
	Price               int64  `gorm:"column:price" json:"price"`                                 // Price in IDR (0 for free plan)
	BillingPeriodDays   int    `gorm:"column:billing_period_days" json:"billing_period_days"`     // e.g. 30 for monthly
	DailyFreeCredits    int    `gorm:"column:daily_free_credits" json:"daily_free_credits"`       // Free credits reset daily
	RoomDurationMinutes int    `gorm:"column:room_duration_minutes" json:"room_duration_minutes"` // Room duration in minutes
	SortOrder           int    `gorm:"column:sort_order;default:0" json:"sort_order"`
	Visibility          bool   `gorm:"column:visibility" json:"visibility"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

type User struct {
	ID                    string     `gorm:"column:id;primaryKey" json:"id"`
	Name                  string     `gorm:"column:name" json:"name"`
	Username              string     `gorm:"column:username" json:"username"`
	Email                 string     `gorm:"column:email" json:"email"`
	Password              string     `gorm:"column:password" json:"-"`
	Credit                int        `gorm:"column:credit" json:"extra_credit"`                             // Extra credits (purchased, never expire)
	FreeCredit            int        `gorm:"column:free_credit;default:0" json:"free_credit"`               // Daily free credits (reset daily)
	FreeCreditResetAt     *time.Time `gorm:"column:free_credit_reset_at" json:"free_credit_reset_at"`       // Last reset timestamp
	SubscriptionPlanID    *string    `gorm:"column:subscription_plan_id" json:"subscription_plan_id"`       // Current subscription plan
	SubscriptionExpiresAt *time.Time `gorm:"column:subscription_expires_at" json:"subscription_expires_at"` // When subscription expires
}

func (User) TableName() string {
	return "users"
}

// TotalCredits returns free credits + extra credits
func (u *User) TotalCredits() int {
	return u.FreeCredit + u.Credit
}

// DeductCredits deducts credits using free credits first, then extra credits.
// Returns false if insufficient total credits.
func (u *User) DeductCredits(amount int) bool {
	if u.TotalCredits() < amount {
		return false
	}
	if u.FreeCredit >= amount {
		u.FreeCredit -= amount
	} else {
		remainder := amount - u.FreeCredit
		u.FreeCredit = 0
		u.Credit -= remainder
	}
	return true
}

// HasActiveSubscription checks if the user has a non-expired paid subscription
func (u *User) HasActiveSubscription() bool {
	return u.SubscriptionPlanID != nil && u.SubscriptionExpiresAt != nil && u.SubscriptionExpiresAt.After(time.Now())
}

// Package represents an extra credit package (one-time purchase)
type Package struct {
	ID            string `gorm:"column:id;primaryKey" json:"id"`
	PackageName   string `gorm:"column:package_name" json:"package_name"`
	PackageDetail []byte `gorm:"column:package_detail" json:"package_detail"`
	Price         int64  `gorm:"column:price" json:"price"` // Price in IDR
	CreditAmount  int    `gorm:"column:credit_amount" json:"credit_amount"`
	Visibility    bool   `gorm:"column:visibility" json:"visibility"`
}

func (Package) TableName() string {
	return "packages"
}

type Room struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	RoomKey     string    `gorm:"column:room_key;uniqueIndex" json:"room_key"`
	RoomCreator string    `gorm:"column:room_creator" json:"room_creator"`
	RoomName    string    `gorm:"column:room_name" json:"room_name"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	RoomMaster  string    `gorm:"column:room_master" json:"room_master"`
	Creator     User      `gorm:"foreignKey:RoomCreator;references:ID" json:"creator"`
	Master      User      `gorm:"foreignKey:RoomMaster;references:ID" json:"master"`
}

func (Room) TableName() string {
	return "rooms"
}

// GetExpiredAt calculates the expiration time based on CreatedAt + room_max_duration
func (r *Room) GetExpiredAt(maxDuration int) time.Time {
	return r.CreatedAt.Add(time.Duration(maxDuration) * time.Minute)
}

// IsExpired checks if the room has expired based on room_max_duration config
func (r *Room) IsExpired(maxDuration int) bool {
	return time.Now().After(r.GetExpiredAt(maxDuration))
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
	ID            string            `gorm:"column:id;primaryKey" json:"id"`
	UserID        string            `gorm:"column:user_id" json:"user_id"`
	PackageID     *string           `gorm:"column:package_id" json:"package_id"`         // For extra credit purchases
	PlanID        *string           `gorm:"column:plan_id" json:"plan_id"`               // For subscription purchases
	Amount        int64             `gorm:"column:amount" json:"amount"`                 // Amount in IDR
	Status        string            `gorm:"column:status" json:"status"`                 // pending, settlement, failed, expired, refunded
	PaymentMethod string            `gorm:"column:payment_method" json:"payment_method"` // flip, free
	TxType        string            `gorm:"column:tx_type" json:"tx_type"`               // extra_credit, subscription
	ExternalID    string            `gorm:"column:external_id" json:"external_id"`       // Flip bill_link_id
	CreatedAt     time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	PaidAt        *time.Time        `gorm:"column:paid_at" json:"paid_at"`
	User          User              `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Package       *Package          `gorm:"foreignKey:PackageID;references:ID" json:"package"`
	Plan          *SubscriptionPlan `gorm:"foreignKey:PlanID;references:ID" json:"plan"`
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
	ConfigRoomMaxDuration     = "room_max_duration"     // in minutes (fallback for free plan)
	ConfigRoomCreationCost    = "room_creation_cost"    // credits required
	ConfigDefaultCredits      = "default_credits"       // extra credits for new users
	ConfigDailyFreeCredits    = "daily_free_credits"    // daily free credits for free plan (default: 5)
	ConfigFlipSecretKey       = "flip_secret_key"       // Flip API Secret Key
	ConfigFlipValidationToken = "flip_validation_token" // Flip Validation Token
	ConfigFlipEnvironment     = "flip_environment"      // "production" or "sandbox"
)

// Transaction type constants
const (
	TxTypeExtraCredit  = "extra_credit"
	TxTypeSubscription = "subscription"
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
	CreditTypePurchase     = "purchase" // Extra credit purchase
	CreditTypeRoomCreation = "room_creation"
	CreditTypeFreeReset    = "free_reset"   // Daily free credit reset
	CreditTypeSubscription = "subscription" // Subscription activation
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
