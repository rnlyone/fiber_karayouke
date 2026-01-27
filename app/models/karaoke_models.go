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
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	RoomKey      string    `gorm:"column:room_key" json:"room_key"`
	RoomCreator  string    `gorm:"column:room_creator" json:"room_creator"`
	RoomName     string    `gorm:"column:room_name" json:"room_name"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	RoomMaster   string    `gorm:"column:room_master" json:"room_master"`
	Creator      User      `gorm:"foreignKey:RoomCreator;references:ID" json:"creator"`
	Master       User      `gorm:"foreignKey:RoomMaster;references:ID" json:"master"`
}

func (Room) TableName() string {
	return "rooms"
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
	ID        string     `gorm:"column:id;primaryKey" json:"id"`
	Name      string     `gorm:"column:name" json:"name"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UserID    *string    `gorm:"column:id_user" json:"user_id"`
	RoomID    string     `gorm:"column:id_room" json:"room_id"`
	Room      Room       `gorm:"foreignKey:RoomID;references:ID" json:"room"`
	User      *User      `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (Guest) TableName() string {
	return "guests"
}
