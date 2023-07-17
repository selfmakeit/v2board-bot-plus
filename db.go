package main

import (
	"fmt"
	"os"
	"math/rand"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type User struct {
	Id                uint
	TelegramId        uint
	Email             string
	Token             string
	IsAdmin           bool
	U                 int64
	D                 int64
	PlanId            int64
	Balance           int64
	TransferEnable    int64
	CommissionBalance int64
	ExpiredAt         int64
	CreatedAt         int64
	PlanName          string
}
type InviteCode struct{
	gorm.Model
	UserId uint
	Code string
	Status int
	Pv int
}
type Invite struct{
	Num int
	InviteUserId uint
	Email string
}

type Plan struct {
	Id   uint
	Name string
	Show int
	Content string
	MonthPrice float64
	OnetimePrice float64
}

type VVBot struct {
	Id             uint `gorm:"primaryKey"`
	UserId         uint `gorm:"unique"`
	TelegramId     uint `gorm:"unique" `
	CheckinTraffic int64
	CheckinAt      int64
	NextAt         int64
}

func (VVBot) TableName() string {
	return "tgbot"
}

var DB *gorm.DB

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.GetString("mysql.username"), config.GetString("mysql.passwd"), config.GetString("mysql.host"), config.GetString("mysql.port"), config.GetString("mysql.database"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "v2_",
			SingularTable: true,
		},
	})
	if err != nil {
		fmt.Printf("连接数据库失败,将不能使用v2board查询功能... \n错误信息: %v", err)
		return nil
	}
	fmt.Println("连接数据库成功")
	if err = db.AutoMigrate(&VVBot{}); err != nil {
		fmt.Printf("创建签到表失败... \n错误信息: %v", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	DB = db
	return db
}

func IsCurrentUserAdmin(tgId int64) bool {
	for _, id := range config.Get("telegram.admins").([]interface{}) {
		if tgId == int64(id.(int)) {
			return true
		}
	}
	var user User
	err := DB.Where("telegram_id = ?", tgId).First(&user).Error
	if err != nil {
		return false
	}
	return user.IsAdmin
}
func QueryPlan(planId int) Plan {
	var plan Plan
	DB.Where("id = ?", planId).First(&plan)
	return plan
}

func QueryUser(tgId int64) User {
	var user User
	DB.Where("telegram_id = ?", tgId).First(&user)
	return user
}

func BindUser(token string, tgId int64) User {
	var user User
	DB.Where("token = ?", token[6:]).First(&user)
	if user.Id <= 0 {
		return user
	}
	if user.TelegramId <= 0 {
		DB.Model(&user).Update("telegram_id", tgId)
	}
	return user
}

func unbindUser(tgId int64) User {
	var user User
	DB.Where("telegram_id = ?", tgId).First(&user)
	if user.Id > 0 {
		DB.Model(&user).Update("telegram_id", nil)
		return user
	}
	return user
}

func CheckinTime(tgId int64) bool {
	var uu VVBot
	DB.Where("telegram_id = ?", tgId).First(&uu)
	if time.Now().Unix() < uu.NextAt {
		return false
	}
	return true
}

func checkinUser(tgId int64) VVBot {
	var user User
	var uu VVBot
	DB.Where("telegram_id = ?", tgId).First(&user)
	DB.Where("telegram_id = ?", tgId).First(&uu)

	r := rand.New(rand.NewSource(time.Now().Unix()))
	b := r.Int63n(config.GetInt64("traffic"))
	CheckIns := b * 1024 * 1024
	T := user.TransferEnable + CheckIns

	if uu.Id <= 0 {
		newUU := VVBot{
			UserId:         user.Id,
			TelegramId:     user.TelegramId,
			CheckinAt:      time.Now().Unix(),
			NextAt:         time.Now().Unix() + 86400,
			CheckinTraffic: 0,
		}
		DB.Create(&newUU)
	}

	DB.Model(&uu).Updates(VVBot{
		CheckinAt:      time.Now().Unix(),
		NextAt:         time.Now().Unix() + 86400,
		CheckinTraffic: CheckIns,
	})
	DB.Model(&user).Update("transfer_enable", T)

	return uu
}
func getIncomeStatistic() string {
	if DB == nil {
		return "未正确配置数据库信息!\n无法使用V2boadr查询功能"
	}
	var r1 float32
	var r2 float32
	var r3 float32
	sql1 := "SELECT SUM(total_amount)/100 as today FROM `v2_order` WHERE to_days(FROM_UNIXTIME(created_at)) = to_days(now()) and status not in(0,2)"
	sql2 := "SELECT SUM(total_amount)/100 as week FROM `v2_order` WHERE date_sub(CURDATE(),INTERVAL 7 DAY) <= DATE(FROM_UNIXTIME(created_at)) and status not in(0,2)"
	sql3 := "SELECT SUM(total_amount)/100 as month FROM `v2_order` WHERE MONTH( FROM_UNIXTIME(created_at))=MONTH(now()) and status not in(0,2)"
	DB.Raw(sql1).Scan(&r1)
	DB.Raw(sql2).Scan(&r2)
	DB.Raw(sql3).Scan(&r3)

	msg := fmt.Sprintf("💲*收益情况:*\n\n今日: %v\n本周: %v\n本月: %v\n", r1, r2, r3)

	return msg
}
func getDayTrafficStatistic(page int) string {
	if DB == nil {
		return "未正确配置数据库信息!\n无法使用V2boadr查询功能"
	}
	var users []User
	sql := fmt.Sprintf("select t.email,u.u,u.d,t.transfer_enable from v2_stat_user  u left join v2_user t on u.user_id =t.id where to_days(FROM_UNIXTIME(u.created_at)) = to_days(now()) and u.record_type ='d' group by t.email order by (u.u+u.d) desc limit %v,10", (page * 10))
	DB.Raw(sql).Scan(&users)
	if len(users) < 1 {
		return "暂无更多信息"
	}
	txt := "📊*今日流量使用排名*\n\n*     排名 | 使用量 | 剩余量  | 邮箱*\n"
	txt += "```    ————————————————————\n"
	for k, v := range users {
		var user = v
		Email := user.Email
		// TransferEnable := ByteSize(user.TransferEnable)
		// U := ByteSize(user.U)
		// D := ByteSize(user.D)
		T := ByteSize(user.D + user.U)
		S := ByteSize(user.TransferEnable - (user.U + user.D))
		txt += fmt.Sprintf("【%v】🚰%s🔋%v \n     📧%v\n", page*10+k+1, T, S, Email)
		txt += "    ————————————————————\n"
	}
	txt += "```\n"

	return txt
}
func getInviteStatistic(page int) string {
	if DB == nil {
		return "未正确配置数据库信息!\n无法使用V2boadr查询功能"
	}
	var invites []Invite
	sql := fmt.Sprintf("select uu.email,count(1)as num,uu.id from v2_user u left join v2_user uu on u.invite_user_id = uu.id where u.invite_user_id is not null group by uu.id order by num desc limit %v,10", (page * 10))
	DB.Raw(sql).Scan(&invites)
	if len(invites) < 1 {
		return "暂无更多信息"
	}
	txt := "📊*邀请人数排名*\n\n*      邮箱 | 邀请人数 *\n"
	txt += "```    ————————————————————\n"
	for k, v := range invites {
		var user = v
		Email := user.Email
		txt += fmt.Sprintf("【%v】📧%s 邀请了%v 人\n", page*10+k+1, Email,user.Num)
		txt += "    ————————————————————\n"
	}
	txt += "```\n"

	return txt
}
func getPlanStatistic(page int) string {
	if DB == nil {
		return "未正确配置数据库信息!\n无法使用V2boadr查询功能"
	}
	type PlanS struct{
		Name string
		Num float64
	}
	var invites []PlanS
	sql := fmt.Sprintf("SELECT (case when p.name is null then '无套餐用户' else p.name end)as name ,count(1) as num from v2_user u  left join v2_plan p on u.plan_id= p.id  group by p.id order by num desc limit %v,10", (page * 10))
	DB.Raw(sql).Scan(&invites)
	if len(invites) < 1 {
		return "暂无更多信息"
	}
	txt := "📊*套餐使用分布*\n\n*      套餐名称 | 使用人数 | 占比*\n"
	txt += "```    ————————————————————\n"
	total :=0.0 
	for _, v := range invites {
		total+=v.Num
	}
	for k, v := range invites {
		txt += fmt.Sprintf("【%v】 %v  => %v人使用,占比 %.1f%% 人\n", page*10+k+1, v.Name,v.Num,v.Num/total*100)
		txt += "    ————————————————————\n"
	}
	txt += "```\n"

	return txt
}
func getMonthTrafficStatistic(page int) string {
	if DB == nil {
		return "未正确配置数据库信息!\n无法使用V2boadr查询功能"
	}
	var users []User
	sql := fmt.Sprintf("select t.email,sum(u.u)as u,sum(u.d)as d,sum(t.transfer_enable)as transfer_enable from v2_stat_user  u left join v2_user t on u.user_id =t.id where MONTH( FROM_UNIXTIME(u.created_at))=MONTH(now()) and u.record_type ='d'  group by t.email order by (sum(u.u)+sum(u.d)) desc limit %v,10", (page * 10))
	DB.Raw(sql).Scan(&users)
	if len(users) < 1 {
		return "暂无更多信息"
	}
	txt := "📊*本月流量使用排名*\n\n*     排名 | 使用量 | 剩余量  | 邮箱*\n"
	txt += "```    ————————————————————\n"
	for k, v := range users {
		var user = v
		Email := user.Email
		// TransferEnable := ByteSize(user.TransferEnable)
		// U := ByteSize(user.U)
		// D := ByteSize(user.D)
		T := ByteSize(user.D + user.U)
		S := ByteSize(user.TransferEnable - (user.U + user.D))
		txt += fmt.Sprintf("【%v】🚰%s🔋%v \n     📧%v\n", page*10+k+1, T, S, Email)
		txt += "    ————————————————————\n"
	}
	txt += "```\n"

	return txt
}
func getIncrementStatistic() string {
	if DB == nil {
		return "⚙️未正确配置数据库信息!\n无法使用V2boadr查询功能"
	}
	type Inc struct {
		Category string
		Count    int
	}
	var r1 []Inc
	var r2 []Inc
	var r3 []Inc
	sql1 := "select tmp.category as category,ifnull(stat.count,0)as count from (select '注册' category union all select '注册并绑定TG')tmp left join (SELECT (case when vu.telegram_id is null then '注册' when vu.telegram_id is not null then '注册并绑定TG' else 'other' end)as category,count(*)as count from v2_user vu where to_days(FROM_UNIXTIME(created_at)) = to_days(now()) group by category) stat on tmp.category=stat.category"
	sql2 := "select tmp.category,ifnull(stat.count,0)as count from (select '注册' category union all select '注册并绑定TG')tmp left join (SELECT (case when vu.telegram_id is null then '注册' when vu.telegram_id is not null then '注册并绑定TG' else 'other' end)as category,count(*)as count from v2_user vu where date_sub(CURDATE(),INTERVAL 7 DAY) <= DATE(FROM_UNIXTIME(created_at)) group by category) stat on tmp.category=stat.category"
	sql3 := "select tmp.category,ifnull(stat.count,0)as count from (select '注册' category union all select '注册并绑定TG')tmp left join (SELECT (case when vu.telegram_id is null then '注册' when vu.telegram_id is not null then '注册并绑定TG' else 'other' end)as category,count(*)as count from v2_user vu where MONTH( FROM_UNIXTIME(created_at))=MONTH(now()) group by category) stat on tmp.category=stat.category"
	// sql3 := "SELECT (case when vu.telegram_id is null then '注册' when vu.telegram_id is not null then '注册并绑定TG' else 'other' end)as category,count(*)as count from v2_user vu where MONTH( FROM_UNIXTIME(created_at))=MONTH(now()) group by category"
	DB.Raw(sql1).Scan(&r1)
	DB.Raw(sql2).Scan(&r2)
	DB.Raw(sql3).Scan(&r3)
	msg := fmt.Sprintf("📈*用户增长情况:*\n\n今日:\n  新增%v：%v\n  新增%v：%v\n\n 七天内: \n  新增%v：%v\n  新增%v：%v\n\n本月: \n  新增%v：%v\n  新增%v：%v\n\n ", r1[0].Category, r1[0].Count, r1[1].Category, r1[1].Count, r2[0].Category, r2[0].Count, r2[1].Category, r2[1].Count, r3[0].Category, r3[0].Count, r3[1].Category, r3[1].Count)

	return msg
}

func getInviteList(telegramId int64)[]Invite{
	var invites []Invite
	sql := fmt.Sprintf(`
	select count(1)as num,invite_user_id from (WITH RECURSIVE cte AS (
		SELECT id, invite_user_id, telegram_id,email
		FROM v2_user
		WHERE telegram_id=%v  -- 找到根节点
		UNION ALL
		SELECT u.id, u.invite_user_id, u.telegram_id,u.email
		FROM v2_user u
		JOIN cte ON u.invite_user_id = cte.id
	  )
	  SELECT id,invite_user_id,telegram_id,email FROM cte)aaa where invite_user_id is not null
	  group by invite_user_id order by num desc;
	`,telegramId)
	DB.Raw(sql).Scan(&invites)
	return invites
}

func getUserById(id uint)User{
	var user User
	DB.First(&user,id)
	return user
}

func getInviteLink(id uint)string{
	var inviteCode InviteCode
	DB.Where("user_id = ? and status =?",id,0).First(&inviteCode)
	if inviteCode.ID == 0{
		inviteCode.UserId = id
		inviteCode.Code = GenerateCode(8)
		DB.Create(&inviteCode)
	}
	return fmt.Sprintf("%v%v",config.GetString("inviteUrl"),inviteCode.Code)
}

func getPlanList()[]Plan{
	var plans []Plan
	//quiery all show != 0
	DB.Where("`show` = ?",1).Find(&plans)
	return plans
}